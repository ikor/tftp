package tftp

import (
	"bytes"
	"fmt"
	"testing"
)

func TestReadFile(t *testing.T) {
	tt := []struct {
		name    string
		p       packetRRQ
		want    packet
		wantErr error
	}{}
	for i, tc := range tt {
		fmt.Printf("%v,%v", i, tc)
	}
}

type thandler struct{}

func (h thandler) ReadFile(f string) Reader {
	return bytes.NewBuffer([]byte("testfile1"))
}

type tsession struct{}
