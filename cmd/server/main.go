package main

import (
	"errors"
	"os"
	"sync"

	"github.com/ikor/tftp/tftp"
)

type file struct {
	name string
	data []byte
}
type handler struct {
	store map[string]*file
	mu    *sync.Mutex
}

func (f *file) Read(b []byte) (int, error) {
	return 0, nil
}
func (f *file) Write() {
	return
}
func (f file) Close() error {
	return nil
}

func (h handler) ReadFile(filename string) (tftp.ReadCloser, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.store[filename]; !ok {
		return nil, os.ErrNotExist
	}
	return h.store[filename], nil
}

func (h handler) WriteFile(filename string) (tftp.WriteCloser, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.store[filename]; ok {
		return nil, errors.New("file already exists")
	}
	f := &file{
		name: filename,
	}
	h.store[f.name] = f
	return nil, nil
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
