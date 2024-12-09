//go:generate mockgen -source ./database.go -destination=./mocks/database.go -package=mock_database
package db

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Database struct with pg pool
type Database struct {
	cluster *pgxpool.Pool
}

// PGX p g x
type PGX interface {
	DBops
	BeginTx(ctx context.Context, options *pgx.TxOptions) error
}

// DBops dp ops
type DBops interface {
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetPool(_ context.Context) *pgxpool.Pool
}

// BeginTx begin tx
func (db Database) BeginTx(_ context.Context, _ *pgx.TxOptions) error {
	return nil
}

// newDatabase create a new database from pg pool
func newDatabase(cluster *pgxpool.Pool) *Database {
	return &Database{cluster: cluster}
}

// GetPool get conn pool
func (db Database) GetPool(_ context.Context) *pgxpool.Pool {
	return db.cluster
}

// Get helper
func (db Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.cluster, dest, query, args...)
}

// Select helper
func (db Database) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.cluster, dest, query, args...)
}

// Exec helper
func (db Database) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.cluster.Exec(ctx, query, args...)
}

// ExecQueryRow helper
func (db Database) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.cluster.QueryRow(ctx, query, args...)
}
