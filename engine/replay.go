package engine

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
)

func replayLog(e *Engine) error {
	file, err := os.Open(e.logPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	for {
		var keyLen uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLen); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		key := make([]byte, keyLen)
		if _, err := io.ReadFull(file, key); err != nil {
			return err
		}

		var valLen uint32
		if err := binary.Read(file, binary.LittleEndian, &valLen); err != nil {
			return err
		}

		value := make([]byte, valLen)
		if _, err := io.ReadFull(file, value); err != nil {
			return err
		}

		var storedChecksum uint32
		if err := binary.Read(file, binary.LittleEndian, &storedChecksum); err != nil {
			return err
		}

		computedChecksum := crc32.ChecksumIEEE(
			append(key, value...),
		)

		if storedChecksum != computedChecksum {
			break
		}

		e.index[string(key)] = string(value)
	}

	return nil
}
