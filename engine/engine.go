package engine

import (
	"errors"
	"os"
)

type Engine struct {
	logPath string
	index   map[string]string
}

func New(logPath string) (*Engine, error) {
	e := &Engine{
		logPath: logPath,
		index:   make(map[string]string),
	}

	// Recover state from disk
	if err := replayLog(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *Engine) Get(key string) (string, error) {
	val, ok := e.index[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
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

	sstablePath := e.logPath + ".sstable"

	if err := WriteSSTable(sstablePath, e.index); err != nil {
		return err
	}

	e.index = make(map[string]string)
	return nil
}
