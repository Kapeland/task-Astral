package storage

import (
	"context"
	"github.com/Kapeland/task-Astral/internal/storage/db"
)

type DbStorage struct {
	DB *db.Database
}

func NewDbStorage(ctx context.Context) (DbStorage, error) {
	var dbStorage DbStorage
	database, err := db.NewDb(ctx)
	if err != nil {
		return DbStorage{}, err
	}
	dbStorage.DB = database
	return dbStorage, nil
}

func (s *DbStorage) Close(ctx context.Context) {
	s.DB.GetPool(ctx).Close()
}
