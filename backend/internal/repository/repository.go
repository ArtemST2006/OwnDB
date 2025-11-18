package repository

import (
	"time"

	"github.com/ArtemST2006/OwnDB/backend/internal/repository/mydb"
)

type SetEnv interface {
	Create(string, time.Time) error
}

type Repository struct {
	SetEnv
}

func NewRepository() *Repository {
	return &Repository{
		SetEnv: mydb.NewSetEnv(),
	}
}
