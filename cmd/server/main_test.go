package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestWriteFile(t *testing.T) {
	h := handler{
		store: make(map[string]*file),
		mu:    &sync.RWMutex{},
	}
	testCases := []struct {
		desc     string
		filename string
		data     []byte
		want     []byte
		wanterr  error
	}{
		{
			desc:     "should write a file to the store",
			filename: "foobar",
			data:     []byte{0, 1, 2, 3, 4},
			want:     []byte{0, 1, 2, 3, 4},
			wanterr:  nil,
		},
	}
	for i, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			goterr := h.WriteFile(tC.filename, tC.data)
			if fmt.Sprintf("%v", goterr) != fmt.Sprintf("%v", tC.wanterr) {
				t.Errorf("%v. goterr: %v, wanterr: %v", i, goterr, tC.wanterr)
			}
			if tC.want != nil {
				trs, _ := h.ReadFile(tC.filename)
				got := make([]byte, len(tC.want))
				trs.Read(got)
				if !bytes.Equal(got, tC.want) {
					t.Errorf("%v. result bytes are not equal to expected: got %v, want %v ", i, got, tC.want)
				}
			}
		})
	}
}

func TestFileOverWrite(t *testing.T) {
	h := handler{
		store: make(map[string]*file),
		mu:    &sync.RWMutex{},
	}
	testCases := []struct {
		desc     string
		filename string
		data     []byte
		want     []byte
		wanterr  error
	}{
		{
			desc:     "should write a file to the store",
			filename: "foobar",
			data:     []byte{0, 1, 2, 3, 4},
			want:     nil,
			wanterr:  errors.New("file already exists"),
		},
	}
	for i, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			_ = h.WriteFile(tC.filename, tC.data)
			goterr := h.WriteFile(tC.filename, tC.data)
			if fmt.Sprintf("%v", goterr) != fmt.Sprintf("%v", tC.wanterr) {
				t.Errorf("%v. goterr: %v, wanterr: %v", i, goterr, tC.wanterr)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	h := handler{
		store: make(map[string]*file),
		mu:    &sync.RWMutex{},
	}
	fname := "foobar"
	data := []byte{0, 1, 2, 3, 4}
	if err := h.WriteFile(fname, data); err != nil {
		t.Fatalf("Failed to write a file: %v", err)
	}

	testCases := []struct {
		desc     string
		filename string
		want     []byte
		wanterr  error
	}{
		{
			desc:     "should return file from the store",
			filename: "foobar",
			want:     []byte{0, 1, 2, 3, 4},
			wanterr:  nil,
		},
		{
			desc:     "should return an error on unknown file",
			filename: "foobar55",
			want:     nil,
			wanterr:  os.ErrNotExist,
		},
	}
	for i, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			trs, goterr := h.ReadFile(tC.filename)
			if fmt.Sprintf("%v", goterr) != fmt.Sprintf("%v", tC.wanterr) {
				t.Errorf("%v. goterr: %v, wanterr: %v", i, goterr, tC.wanterr)
			}
			if tC.want != nil {
				got := make([]byte, len(tC.want))
				trs.Read(got)
				if !bytes.Equal(got, tC.want) {
					t.Errorf("%v. result bytes are not equal to expected: got %v, want %v ", i, got, tC.want)
				}
			}
		})
	}
}

func TestRWFileConcurrent(t *testing.T) {
	h := handler{
		store: make(map[string]*file),
		mu:    &sync.RWMutex{},
	}
	werrCh := make(chan error)
	rerrCh := make(chan error)
	for i := 0; i < 130; i++ {
		fname := fmt.Sprintf("foobar%d", i)
		data := []byte{byte(i)}
		go func(fname string, data []byte, errCh chan error) {
			if err := h.WriteFile(fname, data); err != nil {
				werrCh <- err
			}
			trs, _ := h.ReadFile(fname)
			got := make([]byte, 1)
			trs.Read(got)
			if !bytes.Equal(got, data) {
				errCh <- fmt.Errorf("got %v, want %v", got, data)
			}
		}(fname, data, rerrCh)
	}
	select {
	case werr := <-werrCh:
		t.Errorf("write error: %v", werr)
	case rerr := <-rerrCh:
		t.Errorf("read error: %v", rerr)
	default:
	}
}
