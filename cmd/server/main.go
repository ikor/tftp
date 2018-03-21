package main

import (
	"bytes"
	"errors"
	"fmt"
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

func (h handler) ReadFile(filename string) (tftp.Reader, error) {
	if _, ok := h.store[filename]; !ok {
		return nil, os.ErrNotExist
	}
	t := make([]byte, len(h.store[filename].data))
	copy(t, h.store[filename].data)
	return bytes.NewBuffer(t), nil
}

func (h handler) WriteFile(filename string, data []byte) error {
	defer h.mu.Unlock()
	h.mu.Lock()
	if _, ok := h.store[filename]; ok {
		return errors.New("file already exists")
	}
	t := make([]byte, len(data))
	copy(t, data)
	f := &file{
		name: filename,
		data: t,
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
	port := "1069"
	if v := os.Getenv("TFTP_PORT"); v != "" {
		port = v
	}
	addr := fmt.Sprintf("0.0.0.0:%v", port)
	err := tftp.ListenAndServe(addr, h)
	panic(err)
}
