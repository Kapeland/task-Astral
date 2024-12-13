package models

import (
	"context"
	"github.com/Kapeland/task-Astral/internal/models/structs"
)

type UsersStorager interface {
	CreateUser(ctx context.Context, info structs.RegisterUserInfo) (int, error)
	GetUserByLogin(ctx context.Context, login string) (structs.User, error)
	CheckPassword(ctx context.Context, info structs.AuthUserInfo) (bool, error)
}

// Implement if necessary
