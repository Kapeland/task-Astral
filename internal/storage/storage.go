package storage

import (
	"context"
	"github.com/Kapeland/task-Astral/internal/storage/db"
)

type PostgresStorage struct {
	DB *db.PgDatabase
}

func NewPostgresStorage(ctx context.Context) (PostgresStorage, error) {
	var dbStorage PostgresStorage
	database, err := db.NewPostgres(ctx)
	if err != nil {
		return PostgresStorage{}, err
	}
	dbStorage.DB = database
	return dbStorage, nil
}

func (s *PostgresStorage) Close() {
	s.DB.Close()
}

// type SomeOtherStorage struct {
//	DB *что-то (и не обязательно *sqlx.DB)
// }
// И это что-то должно реализовывать интерфейс из "storage/db" (dbOps)
