package tftp

import (
	"bytes"
	"testing"
)

func testReadPacketRWRQ(t *testing.T, opcode uint16) {
	// opcode|filename|0|mode|0|
	tt := []struct {
		name string
		p    []byte
		err  string
	}{
		{
			name: "invalid packet: invalid mode",
			p:    []byte{opcode, "file1", 0, "foobar", 0},
			err:  "invalid mode",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			readRawPacket(bytes.NewBuffer(tc.p))
		})
	}
}
