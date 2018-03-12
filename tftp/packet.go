package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	opcRRQ uint16 = iota + 1
	opcWRQ
	opcDATA
	opcACK
	opcERR
)

type packet interface {
	Read(b *bytes.Buffer) error
	Write(b *bytes.Buffer) error
}

type packetRWRQ struct {
	mode     string
	filename string
}

type packetRRQ struct {
	packetRWRQ
}

type packetWRQ struct {
	packetRWRQ
}

type packetACK struct {
	blockNum uint16
}

type packetDATA struct {
	blockNum uint16
	data     []byte
}

type packetERR struct {
	errCode uint16
	errMsg  string
}

func readRawPacket(b *bytes.Buffer) (packet, error) {
	var p packet
	var opcode uint16
	if err := binary.Read(b, binary.BigEndian, &opcode); err != nil {
		return nil, err
	}
	switch opcode {
	case opcRRQ:
		p = &packetRRQ{}
	case opcWRQ:
		p = &packetWRQ{}
	case opcACK:
		p = &packetACK{}
	case opcDATA:
		p = &packetDATA{}
	case opcERR:
		p = &packetERR{}
	default:
		return nil, errors.New("could not parse packet type")
	}

	if err := p.Read(b); err != nil {
		return nil, err
	}
	return p, nil
}

func writeRawPacket(p packet, b *bytes.Buffer) error {
	var opcode uint16

	switch p.(type) {
	case *packetRRQ:
		opcode = opcRRQ
	case *packetWRQ:
		opcode = opcWRQ
	case *packetACK:
		opcode = opcACK
	case *packetDATA:
		opcode = opcDATA
	case *packetERR:
		opcode = opcERR
	default:
		return errors.New("could not recognize packet type")
	}

	if err := binary.Write(b, binary.BigEndian, opcode); err != nil {
		return err
	}
	return p.Write(b)
}

func readChunk(b *bytes.Buffer) (string, error) {
	last := b.Bytes()
	p := bytes.IndexByte(last, 0)
	if p < 0 {
		return "", io.ErrUnexpectedEOF
	}
	c := b.Next(p + 1)
	return string(c[:p]), nil
}

func writeChunk(b *bytes.Buffer, s string) error {
	if _, err := b.WriteString(s); err != nil {
		return err
	}
	return b.WriteByte(0)
}

func (p *packetRWRQ) Read(b *bytes.Buffer) error {
	var err error
	p.filename, err = readChunk(b)
	if err != nil {
		return err
	}
	mode, err := readChunk(b)
	if err != nil {
		return err
	}
	if mode != "octet" {
		return errors.New("invalid mode")
	}
	return nil
}

func (p *packetRWRQ) Write(b *bytes.Buffer) error {
	if err := writeChunk(b, p.filename); err != nil {
		return err
	}
	return writeChunk(b, p.mode)
}

func (p *packetACK) Read(b *bytes.Buffer) error {
	return binary.Read(b, binary.BigEndian, &p.blockNum)
}

func (p *packetACK) Write(b *bytes.Buffer) error {
	return binary.Write(b, binary.BigEndian, p.blockNum)
}

func (p *packetDATA) Read(b *bytes.Buffer) error {
	if err := binary.Read(b, binary.BigEndian, &p.blockNum); err != nil {
		return err
	}
	p.data = make([]byte, b.Len())
	_, err := b.Read(p.data)
	return err
}

func (p *packetDATA) Write(b *bytes.Buffer) error {
	if err := binary.Write(b, binary.BigEndian, p.blockNum); err != nil {
		return err
	}
	b.Write(p.data)
	return nil
}

func (p *packetERR) Read(b *bytes.Buffer) error {
	err := binary.Read(b, binary.BigEndian, &p.errCode)
	if err != nil {
		return err
	}
	p.errMsg, err = readChunk(b)
	return err
}

func (p *packetERR) Write(b *bytes.Buffer) error {
	err := binary.Write(b, binary.BigEndian, p.errCode)
	if err != nil {
		return err
	}
	return writeChunk(b, p.errMsg)
}
