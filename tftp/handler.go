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
	ReadFile(f string) (ReadCloser, error)
	WriteFile(f string) (WriteCloser, error)
}

// ReadCloser is an interface for handling read TFTP requests
type ReadCloser interface {
	io.ReadCloser
}

// WriteCloser is an interface for handling write TFTP requests
type WriteCloser interface {
	io.WriteCloser
}

type wireReader interface {
	read(*bytes.Buffer) (packet, error)
}
type wireWriter interface {
	write(packet) error
}
type session struct {
	h     Handler
	raddr *net.UDPAddr
	conn  *net.UDPConn
	wireReader
	wireWriter
}
type ackValidator func(p packet) bool

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
	for blockNum := uint16(1); err == nil; blockNum++ {
		n, err := io.ReadAtLeast(fd, buf, blocksize)
		// ReadAtList will produce 2 errors: io.EOF if n == 0
		// or io.ErrUnexpectedEOF if n < 512
		switch err {
		case nil:
		case io.ErrUnexpectedEOF, io.EOF:
			err = io.EOF
		default:
			s.errorHandler(errNotDefined, err.Error())
			return
		}
		// prepare data packet
		p := &packetDATA{
			blockNum: blockNum,
			data:     buf[:n],
		}
		s.writeAndWait(p, ackVal(blockNum))
	}
}

func (s *session) handleWRQ(p *packetWRQ) {
	return
}

func ackVal(blockNum uint16) ackValidator {
	return func(p packet) bool {
		ack, ok := p.(*packetACK)
		return ok && (ack.blockNum == blockNum)
	}
}

// writeAndWait will send a data packet and wait for an ACK. If no response arrives
// before timeout, it will try to send same packet again until retry limit is reached.
func (s *session) writeAndWait(p packet, v ackValidator) (packet, error) {
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
		ack, err := s.read(bytes.NewBuffer(b[:n]))
		if v(ack) {
			return p, nil
		}
	}

	return nil, errTimeOut
}
