package models

import (
	"context"
	"github.com/Kapeland/task-Astral/internal/models/structs"
)

type UsersStorager interface {
	CreateUser(ctx context.Context, user structs.UserDTO) (int, error)
	GetUserByLogin(ctx context.Context, login string) (structs.UserDTO, error)
}

// Implement if necessary
