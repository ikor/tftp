package tftp

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"
)

const (
	timeout = 5 // in seconds
	retry   = 3
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
	ReadFile(f string, c Conn) (ReadCloser, error)
	WriteFile(f string, c Conn) (WriteCloser, error)
}

// Conn describes current connection's endpoints
type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
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
	read(time.Duration) (packet, error)
}
type wireWriter interface {
	write(packet) error
}
type session struct {
	h       Handler
	c       Conn
	timeout int // seconds
	r       wireReader
	w       wireWriter
}

type packetReader struct {
	ch <-chan []byte
}

func (pr *packetReader) read(timeout time.Duration) (packet, error) {
	select {
	case b := <-pr.ch:
		return readRawPacket(bytes.NewBuffer(b))
	case <-time.After(timeout):
		return nil, errTimeOut
	}
}

type packetWriter struct {
	net.PacketConn
	addr net.Addr
	b    bytes.Buffer
}

func (pw *packetWriter) write(p packet) error {
	pw.b.Reset()
	if err := writeRawPacket(p, &pw.b); err != nil {
		return err
	}
	_, err := pw.PacketConn.WriteTo(pw.b.Bytes(), pw.addr)
	return err
}

func (s *session) errorHandler(code uint16, msg string) error {
	e := packetERR{
		errCode: code,
		errMsg:  msg,
	}
	return s.write(&e)
}

func serve(h Handler, c Conn, r wireReader, w wireWriter) {
	s := &session{h, c, timeout, r, w}
	s.serve()
}

func (s *session) serve() {
	p, err := s.read(0)
	if err != nil {
		s.errorHandler(errNotDefined.errCode, err.Error())
		return
	}
	switch pt := p.(type) {
	case *packetRRQ:
		s.handleRRQ(pt)
	case *packetWRQ:
		s.handleWRQ(pt)
	default:
		s.errorHandler(errIllegalOperation)
	}
	return
}

// writeAndWait will send a packet and wait for response. If no response arrives before timeout,
// it will try to send same packet again until retry limit is reached.
func (s *session) writeWithWait(p packet) (packet, error) {
	for i := 0; i < retry; i++ {
		if err := s.write(p); err != nil {
			return err
		}
	}
}
