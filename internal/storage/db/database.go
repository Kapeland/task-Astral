//go:generate mockgen -source ./database.go -destination=./mocks/database.go -package=mock_database
package db

import (
	"context"
	"database/sql"
)

// DBops dp ops
type DBops interface {
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Begin(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
