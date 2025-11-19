package mydb

import (
	"encoding/binary"
	"os"
)

// getNextID — находит следующий ID (инкремент)
func (c *Central) getNextID() (int64, error) {
	maxID := int64(0)
	file, err := os.Open(c.FileData)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	defer file.Close()

	buf := make([]byte, recordSize)
	for {
		_, err := file.Read(buf)
		if err != nil {
			break
		}
		id := int64(binary.LittleEndian.Uint64(buf[0:8]))
		if id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}
