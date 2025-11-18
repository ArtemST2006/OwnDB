package mydb

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SetEnvDB struct {
	Dir string
}

func NewSetEnv() *SetEnvDB {
	return &SetEnvDB{}
}

func (s *SetEnvDB) Create(name string, time time.Time) error {
	fullPath := filepath.Join(s.Dir, name)
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("база данных с таким именем существует")
	}

	if err := os.Mkdir(fullPath, 0755); err != nil {
		return err
	}

	userFile := filepath.Join(fullPath, name+"_user")
	dataFile := filepath.Join(fullPath, name+"_data")

	file, err := os.Create(userFile)
	if err != nil {
		return err
	}
	file.Close()

	file, err = os.Create(dataFile)
	if err != nil {
		return err
	}
	file.Close()

	return nil
}
