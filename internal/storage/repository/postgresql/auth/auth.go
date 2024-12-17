package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/db"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
)

type Repo struct {
	db db.DBops
}

func New(db db.DBops) *Repo {
	return &Repo{db: db}
}

// CreateUserSecret create user secret
// Returns repository.ErrDuplicateKey or err
func (r *Repo) CreateUserSecret(ctx context.Context, userSecret *structs.UserSecret) error {
	tx, err := r.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx,
		`INSERT INTO auth_schema.users_auth(login, valid_until, token)
				VALUES(:login, :valid_until, :token);`, userSecret)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		if pgErr.Code == "23505" {
			return repository.ErrDuplicateKey
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return err
	}
	return nil
}

// DeleteUserSecret delete user secret
// Returns repository.ErrObjectNotFound or err
func (r *Repo) DeleteUserSecret(ctx context.Context, token string) error {
	tx, err := r.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`DELETE FROM auth_schema.users_auth
	  			WHERE token = $1 ;`, token)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repository.ErrObjectNotFound
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return err
	}

	return nil
}

// GetSecretByLogin get secret
// Returns repository.ErrObjectNotFound or err
func (r *Repo) GetSecretByLogin(ctx context.Context, login string) (*structs.UserSecret, error) {
	userSecret := structs.UserSecret{}

	tx, err := r.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.GetContext(ctx, &userSecret,
		`SELECT login, valid_until, token FROM auth_schema.users_auth
				WHERE login=$1;`, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrObjectNotFound
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return nil, err
	}

	return &userSecret, nil
}

// UpdateUserSecret update secret
// Returns repository.ErrDuplicateKey or err
func (r *Repo) UpdateUserSecret(ctx context.Context, userSecret *structs.UserSecret) error {
	tx, err := r.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE auth_schema.users_auth set
				valid_until = $1 where token=$2;`, userSecret.ValidUntil, userSecret.Token)
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
	}
	if rows == 0 {
		slog.Info(fmt.Sprintf("%d secrects've been updated", rows))
	}
	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return err
	}
	return nil
}

// GetLoginBySecret return login by token.
// Returns repository.ErrObjectNotFound or err
func (r *Repo) GetLoginBySecret(ctx context.Context, secret string) (string, error) {
	login := ""

	tx, err := r.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	err = tx.GetContext(ctx, &login,
		`SELECT login FROM auth_schema.users_auth
				WHERE token=$1;`, secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrObjectNotFound
		}
		return "", err
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return "", err
	}
	return login, nil
}

// GetSecretBySecret get secret
// Returns repository.ErrObjectNotFound or err
func (r *Repo) GetSecretBySecret(ctx context.Context, token string) (*structs.UserSecret, error) {
	userSecret := structs.UserSecret{}

	tx, err := r.db.(*db.PgDatabase).BeginX(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.GetContext(ctx, &userSecret,
		`SELECT login, valid_until, token FROM auth_schema.users_auth
				WHERE token=$1;`, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrObjectNotFound
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		slog.Info("Looks like the context has been closed")
		slog.Error(err.Error())
		return nil, err
	}

	return &userSecret, nil
}
