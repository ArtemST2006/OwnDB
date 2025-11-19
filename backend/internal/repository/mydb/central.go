package mydb

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
	"github.com/xuri/excelize/v2"
)

type Central struct {
	FileData  string
	IndexData string
	index     map[string]int64
	mutex     sync.RWMutex
}

func NewCentral(d_i map[string]int64) *Central {
	return &Central{
		FileData:  "../../db/datadir/data.dat",
		IndexData: "../../db/datadir/data.idx",
		index:     d_i,
	}
}

type DataRecord struct {
	ID      int64
	UserId  int
	Text    [MaxTextLen]byte
	Access  int32
	Name    [32]byte
	Created int64
}

const MaxTextLen = 1024
const recordSize = 8 + 4 + 4 + MaxTextLen + 4 + 32 + 8 // 1148 байт

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
	record.UserId = pb.UserId
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
		record.UserId = int(binary.LittleEndian.Uint32(buf[offset:]))
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

		article := schema.Publish{
			Id:     int(record.ID),
			UserId: record.UserId,
			Text:   text,
			Access: int(record.Access),
			Name:   name,
		}
		allartic.Articles = append(allartic.Articles, article)
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
		record.UserId = int(binary.LittleEndian.Uint32(buf[offset:]))
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
