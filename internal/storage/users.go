package storage

import (
	"context"
	"errors"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/models/structs"
	"github.com/Kapeland/task-Astral/internal/storage/repository"
)

type UsersRepo interface {
	CreateUser(ctx context.Context, user *structs.UserDTO) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*structs.UserDTO, error)
}

type UsersStorage struct {
	usersRepo UsersRepo
}

func NewUsersStorage(usersRepo UsersRepo) UsersStorage {
	return UsersStorage{usersRepo: usersRepo}
}

// CreateUser user
func (s *UsersStorage) CreateUser(ctx context.Context, user structs.UserDTO) (int, error) {
	id, err := s.usersRepo.CreateUser(ctx, &user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return id, models.ErrConflict
		}
		return id, err
	}

	return id, nil
}

// GetUserByLogin user
func (s *UsersStorage) GetUserByLogin(ctx context.Context, login string) (structs.UserDTO, error) {
	user, err := s.usersRepo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			return structs.UserDTO{}, models.ErrNotFound
		}
		return structs.UserDTO{}, err
	}
	return *user, nil
}
