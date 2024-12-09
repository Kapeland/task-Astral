package storage

import (
	"context"
	"errors"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
)

type AuthRepo interface {
	CreateUserSecret(ctx context.Context, secretDTO *structs.UserSecretDTO) error
	GetSecretByLogin(ctx context.Context, login string) (*structs.UserSecretDTO, error)
	UpdateUserSecret(ctx context.Context, secretDTO *structs.UserSecretDTO) error
	DeleteUserSecret(ctx context.Context, token string) error
	GetLoginBySecret(ctx context.Context, secret string) (string, error)
	GetSecretBySecret(ctx context.Context, token string) (*structs.UserSecretDTO, error)
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
func (s *AuthStorage) GetUserSecretByLogin(ctx context.Context, login string) (structs.UserSecretDTO, error) {
	secretDTO, err := s.authRepo.GetSecretByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.UserSecretDTO{}, models.ErrNotFound
		}
		return structs.UserSecretDTO{}, err
	}
	return *secretDTO, nil
}

// CreateUserSecret secret
// Returns models.ErrConflict or err
func (s *AuthStorage) CreateUserSecret(ctx context.Context, secretDTO structs.UserSecretDTO) error {
	err := s.authRepo.CreateUserSecret(ctx, &secretDTO)
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
func (s *AuthStorage) UpdateUserSecret(ctx context.Context, secretDTO structs.UserSecretDTO) error {
	err := s.authRepo.UpdateUserSecret(ctx, &secretDTO)
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
func (s *AuthStorage) GetUserSecretBySecret(ctx context.Context, token string) (structs.UserSecretDTO, error) {
	secretDTO, err := s.authRepo.GetSecretBySecret(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.UserSecretDTO{}, models.ErrNotFound
		}
		return structs.UserSecretDTO{}, err
	}
	return *secretDTO, nil
}
