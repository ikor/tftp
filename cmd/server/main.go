package main

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/ikor/tftp/tftp"
)

type file struct {
	name string
	data []byte
	i    int64
}
type handler struct {
	store map[string]*file
	mu    *sync.Mutex
}

func (f *file) Read(b []byte) (int, error) {
	if f.i >= int64(len(f.data)) {
		return 0, io.EOF
	}
	if len(b) == 0 {
		return 0, nil
	}
	n := copy(b, f.data[f.i:])
	f.i += int64(n)
	return n, nil
}

func (h handler) ReadFile(filename string) (tftp.Reader, error) {
	if _, ok := h.store[filename]; !ok {
		return nil, os.ErrNotExist
	}
	return h.store[filename], nil
}

func (h handler) WriteFile(filename string, data []byte) error {
	defer h.mu.Unlock()
	h.mu.Lock()
	if _, ok := h.store[filename]; ok {
		return errors.New("file already exists")
	}
	f := &file{
		name: filename,
	}
	h.store[f.name] = f
	return nil
}

func (h handler) FileExists(filename string) bool {
	if _, ok := h.store[filename]; !ok {
		return false
	}
	return true
}

func main() {
	s := make(map[string]*file)
	mu := &sync.Mutex{}
	h := handler{
		store: s,
		mu:    mu,
	}
	err := tftp.ListenAndServe("0.0.0.0:1069", h)
	panic(err)
}
