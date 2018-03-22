package tftp

import (
	"bytes"
	"errors"
	"io"
	"net"
	"os"
)

const (
	timeout   = 5 // in seconds
	retry     = 3
	blocksize = 512
)

type tFTPError struct {
	errCode uint16
	errMsg  string
}

var (
	errNotDefined       = tFTPError{0, "Not defined, see error message (if any)."}
	errFileNotFound     = tFTPError{1, "File not found."}
	errAccessViolation  = tFTPError{2, "Access violation."}
	errDiskFull         = tFTPError{3, "Disk full or allocation exceeded."}
	errIllegalOperation = tFTPError{4, "Illegal TFTP operation."}
	errUnknownTID       = tFTPError{5, "Unknown transfer ID."}
	errFileExists       = tFTPError{6, "File already exists."}
	errNoSuchUser       = tFTPError{7, "No such user."}

	errTimeOut = errors.New("timeout reached")
)

// Handler is the main interface that any client using this library needs to implement
type Handler interface {
	ReadFile(f string) (Reader, error)
	WriteFile(f string, data []byte) error
	FileExists(f string) bool
}

// Reader is an interface for handling read TFTP requests
type Reader interface {
	io.Reader
}

type wireReader interface {
	read(*bytes.Buffer) (packet, error)
}

type wireWriter interface {
	write(packet) error
}

type session struct {
	h    Handler
	conn *net.UDPConn
	wireReader
	wireWriter
}

type packetValidator func(p packet) bool

func ackValidator(blockNum uint16) packetValidator {
	return func(p packet) bool {
		ack, ok := p.(*packetACK)
		return ok && (ack.blockNum == blockNum)
	}
}

func dataValidator(blockNum uint16) packetValidator {
	return func(p packet) bool {
		pd, ok := p.(*packetDATA)
		return ok && (pd.blockNum == blockNum)
	}
}

func (s *session) errorHandler(terr tFTPError, msg string) error {
	if msg == "" {
		msg = terr.errMsg
	}
	e := packetERR{terr.errCode, msg}
	return s.write(&e)
}

func (s *session) handleRRQ(p *packetRRQ) {
	fd, err := s.h.ReadFile(p.filename)
	if err != nil {
		switch err {
		case os.ErrNotExist:
			s.errorHandler(errFileNotFound, "")
		default:
			s.errorHandler(errNotDefined, err.Error())
		}
		return
	}

	var buf = make([]byte, blocksize)
	var n int
	var rerr error
	for blockNum := uint16(1); rerr == nil; blockNum++ {
		n, rerr = io.ReadAtLeast(fd, buf, blocksize)
		// ReadAtList will return io.ErrUnexpectedEOF if n < 512,
		// i.e. final packet of the session
		switch rerr {
		case nil:
		case io.EOF, io.ErrUnexpectedEOF:
			rerr = io.ErrUnexpectedEOF
		default:
			s.errorHandler(errNotDefined, rerr.Error())
			return
		}

		// prepare data packet
		p := &packetDATA{
			blockNum: blockNum,
			data:     buf[:n],
		}
		_, werr := s.writeAndWait(p, ackValidator(blockNum))
		if werr != nil {
			return
		}
	}
	return
}

func (s *session) handleWRQ(p *packetWRQ) {
	if ok := s.h.FileExists(p.filename); ok {
		s.errorHandler(errFileExists, "")
		return
	}
	var buf []byte
	var err error
	for blockNum := uint16(0); err == nil; blockNum++ {
		pAck := &packetACK{blockNum}
		writeP, writeErr := s.writeAndWait(pAck, dataValidator(blockNum+1))
		if writeErr != nil {
			return
		}
		v, _ := writeP.(*packetDATA)
		buf = append(buf, v.data...)
		if err != nil {
			return
		}
		if len(v.data) < blocksize {
			// transfer complete
			finAck := &packetACK{blockNum + 1}
			_ = s.write(finAck)
			if err := s.h.WriteFile(p.filename, buf); err != nil {
				s.errorHandler(errNotDefined, err.Error())
			}
			return
		}
	}
	return
}

// writeAndWait will send a data packet and wait for a response. If no response arrives
// before timeout, it will try to send same packet again until retry limit is reached.
func (s *session) writeAndWait(p packet, v packetValidator) (packet, error) {
	b := make([]byte, 1500)
	for i := 0; i < retry; i++ {
		if err := s.write(p); err != nil {
			return nil, err
		}
		n, _, err := s.conn.ReadFromUDP(b)
		if _, ok := err.(net.Error); ok {
			continue
		}

		if err != nil {
			return nil, err
		}
		pr, err := s.read(bytes.NewBuffer(b[:n]))
		if v(pr) {
			return pr, nil
		}
	}

	return nil, errTimeOut
}
