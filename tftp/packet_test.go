package tftp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestReadPacketRRQ(t *testing.T) {
	testReadPacketRWRQ(t, []byte{0x0, byte(opcRRQ)})
}
func TestReadPacketWRQ(t *testing.T) {
	testReadPacketRWRQ(t, []byte{0x0, byte(opcWRQ)})
}

func testReadPacketRWRQ(t *testing.T, opcode []byte) {
	// opcode|filename|0|mode|0|
	var b bytes.Buffer
	// p := &
	tt := []struct {
		name    string
		data    string
		want    packet
		wanterr error
	}{
		{
			name:    "malformed packet: invalid mode",
			data:    "file1\x00foobar\x00",
			want:    nil,
			wanterr: errors.New("invalid mode"),
		},
		{
			name:    "malformed packet: missing mode",
			data:    "file1\x00",
			want:    nil,
			wanterr: io.ErrUnexpectedEOF,
		},
		{
			name:    "malformed packet: missing null",
			data:    "file1\x00octet",
			want:    nil,
			wanterr: io.ErrUnexpectedEOF,
		},
		{
			name:    "valid packet",
			data:    "file1\x00octet\x00",
			want:    nil,
			wanterr: nil,
		},
	}
	for i, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			b.Reset()
			b.Write([]byte{0x0, byte(opcRRQ)})
			b.WriteString(tc.data)
			got, goterr := readRawPacket(&b)

			if got != tc.want && tc.want != nil {
				t.Errorf("%v. got: %#v, want: %#v", i, got, tc.want)
			}
			if fmt.Sprintf("%v", goterr) != fmt.Sprintf("%v", tc.wanterr) {
				t.Errorf("%v. goterr: %v, wanterr: %v", i, goterr, tc.wanterr)
			}
		})
	}
}
