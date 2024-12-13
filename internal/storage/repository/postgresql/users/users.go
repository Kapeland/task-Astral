package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/db"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repo struct {
	db db.DBops
}

func New(db db.DBops) *Repo {
	return &Repo{db: db}
}

// CreateUser create user
func (r *Repo) CreateUser(ctx context.Context, info structs.RegisterUserInfo) (int, error) {
	id := 0
	err := r.db.QueryRow(ctx,
		`INSERT INTO users_schema.users(login, password_hash)
				VALUES($1, crypt($2, gen_salt('bf'))) returning id;`, info.Login, info.Pswd).Scan(&id)

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

// VerifyPassword checks whether the password is correct or no.
func (r *Repo) VerifyPassword(ctx context.Context, info structs.AuthUserInfo) (bool, error) {
	isValid := false
	err := r.db.QueryRow(ctx,
		`SELECT (password_hash = crypt($1, password_hash)) 
    			AS password_match 
				FROM users_schema.users
				WHERE login = $2 ;`, info.Pswd, info.Login).Scan(&isValid)

	switch {
	case err != nil && (errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows)):
		return false, nil
	case err != nil:
		return false, err
	default:
		return isValid, nil
	}
}

// GetUserByLogin get user
func (r *Repo) GetUserByLogin(ctx context.Context, login string) (*structs.User, error) {
	var info structs.User
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
