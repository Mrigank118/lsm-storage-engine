package engine

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("key not found")

type Engine struct {
	logPath  string
	index    map[string]string
	manifest *Manifest
}

func New(logPath string) (*Engine, error) {
	e := &Engine{
		logPath: logPath,
		index:   make(map[string]string),
	}

	if err := replayLog(e); err != nil {
		return nil, err
	}

	dir := filepath.Dir(logPath)
	m, err := OpenManifest(dir)
	if err != nil {
		return nil, err
	}
	e.manifest = m

	return e, nil
}

func (e *Engine) Get(key string) (string, error) {
	if val, ok := e.index[key]; ok {
		return val, nil
	}

	dir := filepath.Dir(e.logPath)

	for _, name := range e.manifest.SSTablesNewestFirst() {
		path := filepath.Join(dir, name)

		r, err := OpenSSTable(path)
		if err != nil {
			continue
		}

		val, ok, err := r.Get(key)
		r.Close()

		if err != nil {
			return "", err
		}
		if ok {
			return val, nil
		}
	}

	return "", ErrNotFound
}

func (e *Engine) Set(key, value string) error {
	file, err := openLog(e.logPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := writeRecord(file, key, value); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	e.index[key] = value
	return nil
}

func (e *Engine) Delete(key string) error {
	file, err := openLog(e.logPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := writeRecord(file, key, ""); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	delete(e.index, key)

	return nil
}

func (e *Engine) Flush() error {
	if len(e.index) == 0 {
		return nil
	}

	dir := filepath.Dir(e.logPath)

	path, err := nextSSTablePath(dir)
	if err != nil {
		return err
	}

	if err := WriteSSTable(path, e.index); err != nil {
		return err
	}

	if err := e.manifest.AddSSTable(filepath.Base(path)); err != nil {
		return err
	}

	e.index = make(map[string]string)
	return nil
}

func (e *Engine) Compact() error {
	dir := filepath.Dir(e.logPath)
	sstables := e.manifest.SSTablesNewestFirst()

	if len(sstables) <= 1 {
		return nil // nothing to compact
	}

	merged := make(map[string]string)

	for _, name := range sstables {
		path := filepath.Join(dir, name)
		r, err := OpenSSTable(path)
		if err != nil {
			return err
		}

		data, err := r.ReadAll()
		r.Close()
		if err != nil {
			return err
		}

		for k, v := range data {
			if _, seen := merged[k]; !seen {
				if v != "" { // tombstone
					merged[k] = v
				}
			}
		}

		r.Close()
	}

	// write new SSTable
	newPath, err := nextSSTablePath(dir)
	if err != nil {
		return err
	}
	if err := WriteSSTable(newPath, merged); err != nil {
		return err
	}

	// update MANIFEST
	base := filepath.Base(newPath)
	if err := e.manifest.AddSSTable(base); err != nil {
		return err
	}

	for _, old := range sstables {
		if err := e.manifest.RemoveSSTable(old); err != nil {
			return err
		}
	}

	// delete old files
	for _, old := range sstables {
		os.Remove(filepath.Join(dir, old))
	}

	return nil
}
