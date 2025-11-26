package mydb

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ArtemST2006/OwnDB/backend/internal/repository/utils"
	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
	"github.com/xuri/excelize/v2"
)

type Central struct {
	FileData  string
	IndexData string
	BackupDir string
	index     map[string]int64
	mutex     sync.RWMutex
}

func NewCentral(d_i map[string]int64) *Central {
	return &Central{
		FileData:  "../../db/datadir/data.dat",
		IndexData: "../../db/datadir/data.idx",
		BackupDir: "../../backup",
		index:     d_i,
	}
}

type DataRecord struct {
	ID      int64
	UserId  int32
	Text    [MaxTextLen]byte
	Access  int32
	Name    [32]byte
	Created int64
}

const MaxTextLen = 1024
const recordSize = 8 + 4 + 4 + MaxTextLen + 4 + 32 + 8 // 1148 байт

func (c *Central) Delete(del schema.Delete) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	file, err := os.OpenFile(c.FileData, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	field := strings.ToLower(del.Title)
	value := del.Value

	// Вспомогательная функция: пометить запись как удалённую по offset
	markDeleted := func(offset int64) error {
		accessBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(accessBuf, 0) // Access = 0
		_, err := file.WriteAt(accessBuf, offset+1036)
		return err
	}

	// Вспомогательная функция: прочитать строку из [N]byte
	toString := func(b []byte) string {
		for i, ch := range b {
			if ch == 0 {
				return string(b[:i])
			}
		}
		return string(b)
	}

	if field == "id" {
		// Удаление по ID — мгновенно
		if offset, exists := c.index[value]; exists {
			return markDeleted(offset)
		}
		return nil
	}

	// Удаление по другому полю — перебираем ВСЕ записи через индекс
	for id, offset := range c.index {
		// Читаем запись по offset
		fmt.Println(id)
		buf := make([]byte, recordSize)
		if _, err := file.ReadAt(buf, offset); err != nil {
			continue // пропускаем, если ошибка
		}

		matches := false

		switch field {
		case "user_id":
			if uid, err := strconv.Atoi(value); err == nil {
				userId := int(binary.LittleEndian.Uint32(buf[8:12]))
				if userId == uid {
					matches = true
				}
			}

		case "access":
			if acc, err := strconv.Atoi(value); err == nil {
				access := int(binary.LittleEndian.Uint32(buf[1036:1040]))
				if access == acc {
					matches = true
				}
			}

		case "name":
			name := toString(buf[1040:1072])
			if name == value {
				matches = true
			}

		case "text":
			text := toString(buf[12:1036])
			if text == value {
				matches = true
			}
		}

		if matches {
			if err := markDeleted(offset); err != nil {
				return err
			}
			// Не удаляем из индекса — ленивое удаление только через Access=0
		}
	}

	return nil
}

func (c *Central) CreateBackup() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Убедимся, что папка бэкапа существует
	if err := os.MkdirAll(c.BackupDir, 0755); err != nil {
		return err
	}

	// Копируем основной файл данных
	if err := c.copyFile(c.FileData, filepath.Join(c.BackupDir, "data.dat")); err != nil {
		return err
	}

	// Копируем индексный файл
	if err := c.copyFile(c.IndexData, filepath.Join(c.BackupDir, "data.idx")); err != nil {
		return err
	}

	return nil
}

func (c *Central) ImportBackup() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Восстанавливаем data.dat
	if err := c.copyFile(filepath.Join(c.BackupDir, "data.dat"), c.FileData); err != nil {
		return err
	}

	// Восстанавливаем data.idx
	if err := c.copyFile(filepath.Join(c.BackupDir, "data.idx"), c.IndexData); err != nil {
		return err
	}

	// Перезагружаем индекс из восстановленного файла
	newIndex := make(map[string]int64)
	if err := utils.ParseSingleIndex(c.IndexData, newIndex); err != nil {
		return err
	}

	c.index = newIndex
	return nil
}

func (c *Central) Publish(pb schema.Publish) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Открываем .dat
	file, err := os.OpenFile(c.FileData, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Генерируем ID (инкремент)
	newID, err := c.getNextID()
	if err != nil {
		return err
	}

	// Создаём запись
	var record DataRecord
	record.ID = newID
	record.UserId = int32(pb.UserId)
	copy(record.Text[:], []byte(pb.Text))
	record.Access = int32(pb.Access)
	copy(record.Name[:], []byte(pb.Name))
	record.Created = time.Now().Unix()

	// Пишем в .dat
	buf := make([]byte, recordSize)
	offset := 0
	binary.LittleEndian.PutUint64(buf[offset:], uint64(record.ID))
	offset += 8
	binary.LittleEndian.PutUint32(buf[offset:], uint32(record.UserId))
	offset += 4

	// Пишем длину текста
	textLen := len(pb.Text)
	if textLen > MaxTextLen {
		textLen = MaxTextLen
	}
	binary.LittleEndian.PutUint32(buf[offset:], uint32(textLen))
	offset += 4

	// Пишем сам текст
	copy(buf[offset:offset+MaxTextLen], []byte(pb.Text)[:textLen])
	offset += MaxTextLen

	binary.LittleEndian.PutUint32(buf[offset:], uint32(record.Access))
	offset += 4
	copy(buf[offset:offset+32], record.Name[:])
	offset += 32
	binary.LittleEndian.PutUint64(buf[offset:], uint64(record.Created))

	// Пишем в .dat
	offsetInFile, err := file.Seek(0, 2) // SEEK_END
	if err != nil {
		return err
	}

	_, err = file.Write(buf)
	if err != nil {
		return err
	}

	// Обновляем индекс в памяти
	c.index[pb.Name] = offsetInFile

	return nil
}

func (c *Central) GetArticles(allartic *schema.AllArticles) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	file, err := os.Open(c.FileData)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, recordSize)
	for {
		_, err := file.Read(buf)
		if err != nil {
			break // конец файла
		}

		var record DataRecord
		offset := 0
		record.ID = int64(binary.LittleEndian.Uint64(buf[offset:]))
		offset += 8
		record.UserId = int32(binary.LittleEndian.Uint32(buf[offset:]))
		offset += 4

		textLen := int(binary.LittleEndian.Uint32(buf[offset:]))
		offset += 4
		record.Text = [MaxTextLen]byte{}
		copy(record.Text[:], buf[offset:offset+MaxTextLen])
		offset += MaxTextLen

		record.Access = int32(binary.LittleEndian.Uint32(buf[offset:]))
		offset += 4
		copy(record.Name[:], buf[offset:offset+32])
		offset += 32
		record.Created = int64(binary.LittleEndian.Uint64(buf[offset:]))

		// Обрезаем текст до реальной длины
		text := string(record.Text[:textLen])

		// Обрезаем имя до реальной длины
		name := string(record.Name[:])
		for i, b := range name {
			if b == 0 {
				name = name[:i]
				break
			}
		}

		if int(record.Access) != 0 {
			article := schema.Publish{
				Id:     int(record.ID),
				UserId: int(record.UserId),
				Text:   text,
				Access: int(record.Access),
				Name:   name,
			}
			allartic.Articles = append(allartic.Articles, article)
		}
	}

	return nil
}

func (c *Central) Import() (*excelize.File, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Создаём новый Excel-файл
	f := excelize.NewFile()
	sheetName := "Data"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)

	// Заголовки
	headers := []string{"Name", "Text", "Access", "UserId"}
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Открываем .dat и читаем все записи
	file, err := os.Open(c.FileData)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rowIndex := 2 // начинаем с 2-й строки
	buf := make([]byte, recordSize)
	for {
		_, err := file.Read(buf)
		if err != nil {
			break // конец файла
		}

		var record DataRecord
		offset := 0
		record.ID = int64(binary.LittleEndian.Uint64(buf[offset:]))
		offset += 8
		record.UserId = int32(binary.LittleEndian.Uint32(buf[offset:]))
		offset += 4

		textLen := int(binary.LittleEndian.Uint32(buf[offset:]))
		offset += 4
		copy(record.Text[:], buf[offset:offset+MaxTextLen])
		offset += MaxTextLen

		record.Access = int32(binary.LittleEndian.Uint32(buf[offset:]))
		offset += 4
		copy(record.Name[:], buf[offset:offset+32])
		offset += 32
		record.Created = int64(binary.LittleEndian.Uint64(buf[offset:]))

		// Обрезаем текст до реальной длины
		text := string(record.Text[:textLen])

		// Обрезаем имя до реальной длины
		name := string(record.Name[:])
		for i, b := range name {
			if b == 0 {
				name = name[:i]
				break
			}
		}

		// Записываем в Excel
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), name)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), text)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), record.Access)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), record.UserId)

		rowIndex++
	}

	return f, nil
}

func (c *Central) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Создаём целевой файл
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
