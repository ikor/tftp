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
	mu    *sync.RWMutex
}

func (h handler) ReadFile(filename string) (tftp.Reader, error) {
	if ok := h.FileExists(filename); !ok {
		return nil, os.ErrNotExist
	}
	h.mu.RLock()
	t := make([]byte, len(h.store[filename].data))
	copy(t, h.store[filename].data)
	h.mu.RUnlock()
	return bytes.NewBuffer(t), nil
}

func (h handler) WriteFile(filename string, data []byte) error {

	if ok := h.FileExists(filename); ok {
		return errors.New("file already exists")
	}
	t := make([]byte, len(data))
	copy(t, data)
	f := &file{
		name: filename,
		data: t,
	}
	h.mu.Lock()
	h.store[f.name] = f
	h.mu.Unlock()
	return nil
}

func (h handler) FileExists(filename string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.store[filename]
	return ok
}

func main() {
	s := make(map[string]*file)
	mu := &sync.RWMutex{}
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
