package mydb

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
)

type AuthDB struct {
	FileUser  string
	IndexUser string
	index     map[string]int64 // name -> offset
	mutex     sync.RWMutex
}

func NewAuth(u_i map[string]int64) *AuthDB {
	return &AuthDB{
		FileUser:  "../../db/user.dat",
		IndexUser: "../../db/user.idx",
		index:     u_i,
	}
}

// UserRecord — структура записи (фиксированная длина)
type UserRecord struct {
	ID       int64
	Name     [32]byte
	Password [32]byte
	Access   int32
	Created  int64
}

const recordSizeUser = 8 + 32 + 32 + 4 + 8 // 84 байта

func (a *AuthDB) AddUser(sign schema.SignUp) (int64, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.index[sign.Name]; exists {
		return 0, fmt.Errorf("user with name '%s' already exists", sign.Name)
	}

	// Открываем .dat
	file, err := os.OpenFile(a.FileUser, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Генерируем ID (инкремент)
	newID := time.Now().UnixNano()

	// Создаём запись
	var record UserRecord
	record.ID = newID
	copy(record.Name[:], []byte(sign.Name))
	copy(record.Password[:], []byte(sign.Password))
	record.Access = int32(sign.Access)
	record.Created = time.Now().Unix()

	// Пишем в .dat
	buf := make([]byte, recordSizeUser)
	offset := 0
	binary.LittleEndian.PutUint64(buf[offset:], uint64(record.ID))
	offset += 8
	copy(buf[offset:offset+32], record.Name[:])
	offset += 32
	copy(buf[offset:offset+32], record.Password[:])
	offset += 32
	binary.LittleEndian.PutUint32(buf[offset:], uint32(record.Access))
	offset += 4
	binary.LittleEndian.PutUint64(buf[offset:], uint64(record.Created))

	// Пишем в .dat
	offsetInFile, err := file.Seek(0, 2) // SEEK_END
	if err != nil {
		return 0, err
	}

	_, err = file.Write(buf)
	if err != nil {
		return 0, err
	}

	// Обновляем индекс в памяти
	a.index[sign.Name] = offsetInFile

	return newID, nil
}

func (a *AuthDB) Login(sign schema.SignIn) (int64, int32, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Ищем смещение по имени
	offset, exists := a.index[sign.Name]
	if !exists {
		return 0, 0, fmt.Errorf("user not found")
	}

	// Открываем .dat
	file, err := os.Open(a.FileUser)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	// Читаем запись по смещению
	var record UserRecord
	buf := make([]byte, recordSizeUser)
	_, err = file.Seek(offset, 0)
	if err != nil {
		return 0, 0, err
	}
	_, err = file.Read(buf)
	if err != nil {
		return 0, 0, err
	}

	// Десериализуем
	record.ID = int64(binary.LittleEndian.Uint64(buf[0:8]))
	copy(record.Name[:], buf[8:40])
	copy(record.Password[:], buf[40:72])
	record.Access = int32(binary.LittleEndian.Uint32(buf[72:76]))
	record.Created = int64(binary.LittleEndian.Uint64(buf[76:84]))

	// Проверяем пароль
	password := string(record.Password[:])
	for i, b := range password {
		if b == 0 {
			password = password[:i]
			break
		}
	}

	if password == sign.Password {
		return record.ID, record.Access, nil
	}

	return 0, 0, fmt.Errorf("password incorrect")
}
