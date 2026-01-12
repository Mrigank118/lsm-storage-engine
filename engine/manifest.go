package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Manifest struct {
	path     string
	sstables []string
}

func OpenManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, "MANIFEST")

	m := &Manifest{
		path:     path,
		sstables: []string{},
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			break // stop on malformed entry
		}

		switch parts[0] {
		case "ADD":
			m.sstables = append(m.sstables, parts[1])
		case "REMOVE":
			name := parts[1]
			filtered := m.sstables[:0]
			for _, s := range m.sstables {
				if s != name {
					filtered = append(filtered, s)
				}
			}
			m.sstables = filtered
		}

	}

	return m, nil
}

func (m *Manifest) AddSSTable(name string) error {
	f, err := os.OpenFile(m.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "ADD %s\n", name); err != nil {
		return err
	}

	return f.Sync()
}

func (m *Manifest) RemoveSSTable(name string) error {
	f, err := os.OpenFile(m.path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "REMOVE %s\n", name); err != nil {
		return err
	}
	return f.Sync()
}

func (m *Manifest) SSTablesNewestFirst() []string {
	out := make([]string, 0, len(m.sstables))
	for i := len(m.sstables) - 1; i >= 0; i-- {
		out = append(out, m.sstables[i])
	}
	return out
}
