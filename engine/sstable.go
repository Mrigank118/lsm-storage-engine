package engine

import (
	"encoding/binary"
	"os"
	"sort"
)

type SSTableWriter struct {
	file *os.File
}

func NewSSTableWriter(path string) (*SSTableWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &SSTableWriter{file: f}, nil
}

func (w *SSTableWriter) Write(key, value string) error {
	keyBytes := []byte(key)
	valBytes := []byte(value)

	if err := binary.Write(w.file, binary.BigEndian, uint32(len(keyBytes))); err != nil {
		return err
	}
	if _, err := w.file.Write(keyBytes); err != nil {
		return err
	}

	if err := binary.Write(w.file, binary.BigEndian, uint32(len(valBytes))); err != nil {
		return err
	}
	if _, err := w.file.Write(valBytes); err != nil {
		return err
	}

	return nil
}

func (w *SSTableWriter) Close() error {
	if err := w.file.Sync(); err != nil {
		return err
	}
	return w.file.Close()
}

func WriteSSTable(path string, mem map[string]string) error {
	keys := make([]string, 0, len(mem))
	for k := range mem {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	writer, err := NewSSTableWriter(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	for _, k := range keys {
		if err := writer.Write(k, mem[k]); err != nil {
			return err
		}
	}
	return nil
}
