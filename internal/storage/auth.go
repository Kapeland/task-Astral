package storage

import (
	"context"
	"errors"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
)

type AuthRepo interface {
	CreateUserSecret(ctx context.Context, userSecret *structs.UserSecret) error
	GetSecretByLogin(ctx context.Context, login string) (*structs.UserSecret, error)
	UpdateUserSecret(ctx context.Context, userSecret *structs.UserSecret) error
	DeleteUserSecret(ctx context.Context, token string) error
	GetLoginBySecret(ctx context.Context, secret string) (string, error)
	GetSecretBySecret(ctx context.Context, token string) (*structs.UserSecret, error)
}

type AuthStorage struct {
	authRepo AuthRepo
}

func NewAuthStorage(authRepo AuthRepo) AuthStorage {
	return AuthStorage{authRepo: authRepo}
}

// GetUserLoginBySecret secret.
// Returns models.ErrNotFound or err
func (s *AuthStorage) GetUserLoginBySecret(ctx context.Context, secret string) (string, error) {
	login, err := s.authRepo.GetLoginBySecret(ctx, secret)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return "", models.ErrNotFound
		}
		return "", err
	}
	return login, nil
}

// GetUserSecretByLogin secret
// Returns models.ErrNotFound or err
func (s *AuthStorage) GetUserSecretByLogin(ctx context.Context, login string) (structs.UserSecret, error) {
	userSecret, err := s.authRepo.GetSecretByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.UserSecret{}, models.ErrNotFound
		}
		return structs.UserSecret{}, err
	}
	return *userSecret, nil
}

// CreateUserSecret secret
// Returns models.ErrConflict or err
func (s *AuthStorage) CreateUserSecret(ctx context.Context, userSecret structs.UserSecret) error {
	err := s.authRepo.CreateUserSecret(ctx, &userSecret)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return models.ErrConflict
		}
		return err
	}
	return nil
}

// DeleteUserSecret secret
// Returns models.ErrNotFound or err
func (s *AuthStorage) DeleteUserSecret(ctx context.Context, token string) error {
	err := s.authRepo.DeleteUserSecret(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return models.ErrNotFound
		}
		return err
	}
	return nil
}

// UpdateUserSecret secret
// Returns models.ErrConflict or err
func (s *AuthStorage) UpdateUserSecret(ctx context.Context, userSecret structs.UserSecret) error {
	err := s.authRepo.UpdateUserSecret(ctx, &userSecret)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return models.ErrConflict
		}
		return err
	}
	return nil
}

// GetUserSecretBySecret secret.
// Returns models.ErrNotFound or err
func (s *AuthStorage) GetUserSecretBySecret(ctx context.Context, token string) (structs.UserSecret, error) {
	userSecret, err := s.authRepo.GetSecretBySecret(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.UserSecret{}, models.ErrNotFound
		}
		return structs.UserSecret{}, err
	}
	return *userSecret, nil
}
