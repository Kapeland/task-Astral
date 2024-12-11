package db

import (
	"context"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/utils/config"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// NewDb create new db
func NewDb(ctx context.Context) (*Database, error) {
	lgr := logger.GetLogger()

	dsn := generateDsn()
	var err error

	for i := 0; i < 10; i++ {
		pool, err1 := pgxpool.Connect(ctx, dsn)
		if err1 == nil {
			lgr.Info("Successfully connected to database", "client", "NewDb", "pgxpool.Connect")
			lgr.Info(dsn, "", "", "")
			return newDatabase(pool), nil
		}
		lgr.Error(err1.Error(), "client", "NewDb", "pgxpool.Connect")

		err = err1
		time.Sleep(time.Second * 1)
	}
	return nil, err
}

func generateDsn() string {
	cfg := config.GetConfig()

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name)
}
