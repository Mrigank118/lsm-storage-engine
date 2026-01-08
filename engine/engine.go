package engine

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("key not found")

type Engine struct {
	logPath string
	index   map[string]string
}

func New(logPath string) (*Engine, error) {
	e := &Engine{
		logPath: logPath,
		index:   make(map[string]string),
	}

	if err := replayLog(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Engine) Get(key string) (string, error) {
	if val, ok := e.index[key]; ok {
		return val, nil
	}

	dir := filepath.Dir(e.logPath)

	tables, err := listSSTables(dir)
	if err != nil {
		return "", err
	}

	for _, path := range tables {
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

func (e *Engine) Compact() error {
	tmpPath := e.logPath + ".compact"

	// 1. Create new compacted log
	file, err := os.OpenFile(
		tmpPath,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	// 2. Write current state only
	for key, value := range e.index {
		if err := writeRecord(file, key, value); err != nil {
			file.Close()
			return err
		}
	}

	// 3. Ensure durability
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	file.Close()

	// 4. Atomically replace old log
	return os.Rename(tmpPath, e.logPath)
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

	e.index = make(map[string]string)
	return nil
}
