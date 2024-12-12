package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/db"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
	"github.com/jackc/pgx/v5"
	"log/slog"
)

type Repo struct {
	db db.DBops
}

func New(db db.DBops) *Repo {
	return &Repo{db: db}
}

// CreateUserSecret create user secret
// Returns epository.ErrDuplicateKey or err
func (r *Repo) CreateUserSecret(ctx context.Context, secretDTO *structs.UserSecretDTO) error {

	res, err := r.db.(*db.PgDatabase).NamedExec(ctx,
		`INSERT INTO auth_schema.users_auth(login, valid_until, token)
				VALUES(:login, :valid_until, :token);`, secretDTO)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
	}
	if rows == 0 {
	}
	return nil
}

// DeleteUserSecret delete user secret
// Returns repository.ErrObjectNotFound or err
func (r *Repo) DeleteUserSecret(ctx context.Context, token string) error {
	res, err := r.db.Exec(ctx,
		`DELETE FROM auth_schema.users_auth
	  			WHERE token = $1 ;`, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return repository.ErrObjectNotFound
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	if rows == 0 {
		return repository.ErrObjectNotFound
	}
	return nil
}

// GetSecretByLogin get secret
// Returns repository.ErrObjectNotFound or err
func (r *Repo) GetSecretByLogin(ctx context.Context, login string) (*structs.UserSecretDTO, error) {
	secretDTO := structs.UserSecretDTO{}
	err := r.db.Get(ctx, &secretDTO,
		`SELECT login, valid_until, token FROM auth_schema.users_auth
				WHERE login=$1;`, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrObjectNotFound
		}
		return nil, err
	}
	return &secretDTO, nil
}

// UpdateUserSecret update secret
// Returns repository.ErrDuplicateKey or err
func (r *Repo) UpdateUserSecret(ctx context.Context, secretDTO *structs.UserSecretDTO) error {
	res, err := r.db.Exec(ctx,
		`UPDATE auth_schema.users_auth set
				valid_until = $1 where token=$2;`, secretDTO.ValidUntil, secretDTO.Token)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
	}
	if rows == 0 {
	}
	return nil
}

// GetLoginBySecret return login by token.
// Returns repository.ErrObjectNotFound or err
func (r *Repo) GetLoginBySecret(ctx context.Context, secret string) (string, error) {
	login := ""
	err := r.db.Get(ctx, &login,
		`SELECT login FROM auth_schema.users_auth
				WHERE token=$1;`, secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrObjectNotFound
		}
		return "", err
	}
	return login, nil
}

// GetSecretBySecret get secret
// Returns repository.ErrObjectNotFound or err
func (r *Repo) GetSecretBySecret(ctx context.Context, token string) (*structs.UserSecretDTO, error) {
	secretDTO := structs.UserSecretDTO{}
	err := r.db.Get(ctx, &secretDTO,
		`SELECT login, valid_until, token FROM auth_schema.users_auth
				WHERE token=$1;`, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrObjectNotFound
		}
		return nil, err
	}
	return &secretDTO, nil
}
