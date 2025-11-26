package repository

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ArtemST2006/OwnDB/backend/internal/repository/utils"
)

func Settings() bool {
	fullPath := "../../db"
	if _, err := os.Stat(fullPath); err == nil {
		fmt.Println("db has already created")
		return true
	}

	err := os.Mkdir(fullPath, 0755)
	if err != nil {
		return false
	}

	dataDir := filepath.Join(fullPath, "datadir")
	err = os.Mkdir(dataDir, 0755)
	if err != nil {
		return false
	}

	// Файлы данных
	dataDat := filepath.Join(dataDir, "data.dat")
	file, err := os.Create(dataDat)
	if err != nil {
		return false
	}
	file.Close()

	userDat := filepath.Join(fullPath, "user.dat")
	file, err = os.Create(userDat)
	if err != nil {
		return false
	}
	file.Close()

	dataIdx := filepath.Join(dataDir, "data.idx")
	file, err = os.Create(dataIdx)
	if err != nil {
		return false
	}
	file.Close()

	userIdx := filepath.Join(fullPath, "user.idx")
	file, err = os.Create(userIdx)
	if err != nil {
		return false
	}
	file.Close()

	return true
}

func ParseToIndex() (map[string]int64, map[string]int64) {
	userIndex := make(map[string]int64)
	dataIndex := make(map[string]int64)

	// Пути к файлам
	userIdxPath := "../../db/user.idx"
	dataIdxPath := "../../db/datadir/data.idx"

	// Парсим user.idx
	utils.ParseSingleIndex(userIdxPath, userIndex)

	// Парсим data.idx
	utils.ParseSingleIndex(dataIdxPath, dataIndex)

	return userIndex, dataIndex
}

func (r *Repository) Flush() error {
	if err := flushIndex(r.PathIndexUser, r.User); err != nil {
		return fmt.Errorf("failed to flush user index: %w", err)
	}

	if err := flushIndex(r.PathIndexData, r.Data); err != nil {
		return fmt.Errorf("failed to flush data index: %w", err)
	}

	return nil
}

func flushIndex(indexPath string, index map[string]int64) error {
	file, err := os.OpenFile(indexPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for name, offset := range index {
		buf := make([]byte, 16+32)
		binary.LittleEndian.PutUint64(buf[0:8], 0)
		binary.LittleEndian.PutUint64(buf[8:16], uint64(offset))
		copy(buf[16:16+32], []byte(name))

		_, err := file.Write(buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) Clear() error {
	for k := range r.Data {
		delete(r.Data, k)
	}

	for k := range r.User {
		delete(r.User, k)
	}

	return nil
}
