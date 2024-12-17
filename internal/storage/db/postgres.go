package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/utils/config"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// NewPostgres create new db
func NewPostgres(ctx context.Context) (*PgDatabase, error) {
	lgr := logger.GetLogger()

	dsn := generateDsn()
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		lgr.Error("failed to create database connection", "postgres", "NewPostgres", "sqlx.Open")

		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		lgr.Error("failed ping the database", "postgres", "NewPostgres", "PingContext")

		return nil, err
	}

	return &PgDatabase{db}, nil
}

func generateDsn() string {
	cfg := config.GetConfig()

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SslMode)
}

// PgDatabase struct with pg pool
type PgDatabase struct {
	db *sqlx.DB
}

// Get helper
func (db PgDatabase) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.db.GetContext(ctx, dest, query, args...)
}

// Select helper
func (db PgDatabase) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.db.SelectContext(ctx, dest, query, args...)
}

// Exec helper
func (db PgDatabase) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

// NamedExec helper
func (db PgDatabase) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return db.db.NamedExecContext(ctx, query, arg)
}

// QueryRow helper
func (db PgDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.db.QueryRowContext(ctx, query, args...)
}

// QueryRowx helper
func (db PgDatabase) QueryRowx(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return db.db.QueryRowxContext(ctx, query, args...)
}

// NamedQuery helper
func (db PgDatabase) NamedQuery(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	return db.db.NamedQueryContext(ctx, query, arg)
}

// Close closes db
func (db PgDatabase) Close() {
	db.db.Close()
}

// GetDB returns undrlying db
func (db PgDatabase) GetDB() *sqlx.DB {
	return db.db
}

// Begin begins transaction
func (db PgDatabase) Begin(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.db.BeginTx(ctx, opts)
}

// BeginX begins transaction
func (db PgDatabase) BeginX(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return db.db.BeginTxx(ctx, opts)
}
