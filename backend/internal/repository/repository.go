package repository

import (
	"github.com/ArtemST2006/OwnDB/backend/internal/repository/mydb"
	"github.com/ArtemST2006/OwnDB/backend/internal/schema"
	"github.com/xuri/excelize/v2"
)

type Central interface {
	Publish(schema.Publish) error
	GetArticles(*schema.AllArticles) error
	Import() (*excelize.File, error)
	ImportBackup() error
	CreateBackup() error
	Delete(schema.Delete) error
}

type Authorization interface {
	AddUser(schema.SignUp) (int64, error)
	Login(schema.SignIn) (int64, int32, error)
}

type Repository struct {
	Authorization
	Central
	Data          map[string]int64
	User          map[string]int64
	PathUser      string
	PathIndexUser string
	PathData      string
	PathIndexData string
}

func NewRepository(u_i, d_i map[string]int64) *Repository {
	return &Repository{
		Authorization: mydb.NewAuth(u_i),
		Central:       mydb.NewCentral(d_i),
		Data:          d_i,
		User:          u_i,
		PathUser:      "../../db/user.dat",
		PathIndexUser: "../../db/user.idx",
		PathData:      "../../db/datadir/data.dat",
		PathIndexData: "../../db/datadir/data.idx",
	}
}
