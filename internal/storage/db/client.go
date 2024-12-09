package db

import (
	"context"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/utils/configer"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// NewDb create new db
func NewDb(ctx context.Context) (*Database, error) {
	dsn, err := generateDsn()
	if err != nil {
		return nil, err
	}

	for i := 0; i < 10; i++ {
		pool, err1 := pgxpool.Connect(ctx, dsn)
		if err1 == nil {
			log.Println("Successfully connected to database")
			return newDatabase(pool), nil
		}
		log.Println(err1)
		err = err1
		time.Sleep(time.Second * 1)
	}
	return nil, err
}

func generateDsn() (string, error) {
	cfg, err := configer.GetConfig()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name), nil
}
