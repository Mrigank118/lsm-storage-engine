package engine

import (
	"encoding/binary"
	"os"
)

func openLog(path string) (*os.File, error) {
	return os.OpenFile(
		path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
}

func writeRecord(file *os.File, key, value string) error {
	keyLen := uint32(len(key))
	valLen := uint32(len(value))

	if err := binary.Write(file, binary.LittleEndian, keyLen); err != nil {
		return err
	}
	if _, err := file.Write([]byte(key)); err != nil {
		return err
	}

	if err := binary.Write(file, binary.LittleEndian, valLen); err != nil {
		return err
	}
	if _, err := file.Write([]byte(value)); err != nil {
		return err
	}

	return nil
}
