package tftp

import (
	"bytes"
	"net"
	"time"
)

type server struct {
	conn    *net.UDPConn
	handler Handler
}

type packetReader struct {
}

func (pr *packetReader) read(b *bytes.Buffer) (packet, error) {
	p, err := readRawPacket(b)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type packetWriter struct {
	c    *net.UDPConn
	addr net.Addr
	b    bytes.Buffer
}

func (pw *packetWriter) write(p packet) error {
	pw.b.Reset()
	if err := writeRawPacket(p, &pw.b); err != nil {
		return err
	}
	_, err := pw.c.Write(pw.b.Bytes())
	return err
}

func writeError(pw *packetWriter, err tFTPError, msg string) {
	e := &packetERR{err.errCode, msg}
	pw.write(e)
}

func serve(p packet, raddr *net.UDPAddr, h Handler) {
	laddr, err := net.ResolveUDPAddr("udp4", ":0")
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp4", laddr, raddr)
	if err != nil {
		return
	}
	conn.SetDeadline(time.Now().Add(timeout * time.Second))

	spr := &packetReader{}
	spw := &packetWriter{
		c: conn,
	}

	s := session{
		h:          h,
		conn:       conn,
		raddr:      raddr,
		wireReader: spr,
		wireWriter: spw,
	}

	switch pt := p.(type) {
	case *packetRRQ:
		s.handleRRQ(pt)
	case *packetWRQ:
		s.handleWRQ(pt)
	default:
		// writeError(pw, errIllegalOperation, "")
	}
}

// ListenAndServe listens for incoming UPD connections on specified
// host:port and handles them using specified Handler
func ListenAndServe(addr string, h Handler) error {

	laddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}
	c, err := net.ListenUDP("udp4", laddr)
	defer c.Close()

	if err != nil {
		return err
	}

	b := make([]byte, 1500) // Ethernet v2 MTU, todo convert to sync.pool

	for {
		n, raddr, err := c.ReadFromUDP(b)
		if err != nil {
			return err
		}
		pr := &packetReader{}
		p, err := pr.read(bytes.NewBuffer(b[:n]))
		go serve(p, raddr, h)
	}
}
