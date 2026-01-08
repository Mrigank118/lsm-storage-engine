package engine

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
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

type SSTableReader struct {
	file *os.File
}

func OpenSSTable(path string) (*SSTableReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &SSTableReader{file: f}, nil
}

func (r *SSTableReader) Get(target string) (string, bool, error) {
	_, err := r.file.Seek(0, 0)
	if err != nil {
		return "", false, err
	}

	for {
		var keyLen uint32
		if err := binary.Read(r.file, binary.BigEndian, &keyLen); err != nil {
			return "", false, nil
		}

		key := make([]byte, keyLen)
		if _, err := r.file.Read(key); err != nil {
			return "", false, err
		}

		var valLen uint32
		if err := binary.Read(r.file, binary.BigEndian, &valLen); err != nil {
			return "", false, err
		}

		val := make([]byte, valLen)
		if _, err := r.file.Read(val); err != nil {
			return "", false, err
		}

		if string(key) == target {
			return string(val), true, nil
		}
	}
}

func (r *SSTableReader) Close() error {
	return r.file.Close()
}

func nextSSTablePath(dir string) (string, error) {
	pattern := filepath.Join(dir, "sstable-*.db")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	max := 0
	for _, f := range files {
		base := filepath.Base(f)
		var gen int
		fmt.Sscanf(base, "sstable-%06d.db", &gen)
		if gen > max {
			max = gen
		}
	}

	return filepath.Join(dir, fmt.Sprintf("sstable-%06d.db", max+1)), nil
}

func listSSTables(dir string) ([]string, error) {
	pattern := filepath.Join(dir, "sstable-*.db")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	return files, nil
}
