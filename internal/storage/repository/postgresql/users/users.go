package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/db"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type Repo struct {
	db db.DBops
}

func New(db db.DBops) *Repo {
	return &Repo{db: db}
}

// CreateUser create user
func (r *Repo) CreateUser(ctx context.Context, user *structs.UserDTO) (int, error) {
	id := 0
	err := r.db.ExecQueryRow(ctx,
		`INSERT INTO users_schema.users(login, password_hash)
				VALUES($1,$2) returning id;`, user.Login, user.PasswordHash).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		if pgErr.Code == "23505" {
			return id, repository.ErrDuplicateKey
		}
		return id, err
	}
	return id, nil
}

// GetUserByLogin get user
func (r *Repo) GetUserByLogin(ctx context.Context, login string) (*structs.UserDTO, error) {
	var info structs.UserDTO
	err := r.db.Get(ctx, &info,
		`SELECT id, login, password_hash FROM users_schema.users WHERE login=$1;`, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrObjectNotFound
		}
		return nil, err
	}
	return &info, nil
}
