package storage

import (
	"context"
	"errors"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
)

type UsersRepo interface {
	CreateUser(ctx context.Context, info structs.RegisterUserInfo) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*structs.User, error)
	VerifyPassword(ctx context.Context, info structs.AuthUserInfo) (bool, error)
}

type UsersStorage struct {
	usersRepo UsersRepo
}

func NewUsersStorage(usersRepo UsersRepo) UsersStorage {
	return UsersStorage{usersRepo: usersRepo}
}

// CreateUser user
func (s *UsersStorage) CreateUser(ctx context.Context, info structs.RegisterUserInfo) (int, error) {
	id, err := s.usersRepo.CreateUser(ctx, info)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return id, models.ErrConflict
		}
		return id, err
	}

	return id, nil
}

// GetUserByLogin user
func (s *UsersStorage) GetUserByLogin(ctx context.Context, login string) (structs.User, error) {
	user, err := s.usersRepo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.User{}, models.ErrNotFound
		}
		return structs.User{}, err
	}
	return *user, nil
}

// CheckPassword user
func (s *UsersStorage) CheckPassword(ctx context.Context, info structs.AuthUserInfo) (bool, error) {
	ok, err := s.usersRepo.VerifyPassword(ctx, info)
	if err != nil {
		return false, err
	}
	return ok, nil
}
