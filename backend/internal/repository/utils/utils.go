package utils

import (
	"encoding/binary"
	"os"
)

// parseSingleIndex — читает один .idx файл и заполняет переданную мапу
func ParseSingleIndex(filePath string, index map[string]int64) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Файл не существует — это нормально для пустой БД
			return nil
		}
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	size := stat.Size()
	if size == 0 {
		return nil // пустой индекс
	}

	buf := make([]byte, size)
	_, err = file.Read(buf)
	if err != nil {
		return err
	}

	// Каждая запись: [8 байт: ID] + [8 байт: offset] + [32 байта: имя]
	for i := 0; i < len(buf); i += 16 + 32 { // 8+8+32 = 48
		if i+48 > len(buf) {
			break // недостаточно данных
		}

		// Извлекаем имя
		nameBytes := buf[i+16 : i+16+32]
		name := string(nameBytes)
		// Обрезаем нулевые символы
		for j, b := range name {
			if b == 0 {
				name = name[:j]
				break
			}
		}

		// Извлекаем offset
		offset := int64(binary.LittleEndian.Uint64(buf[i+8 : i+16]))

		index[name] = offset
	}

	return nil
}
