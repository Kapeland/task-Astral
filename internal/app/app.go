package app

import (
	"context"
	"fmt"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/services"
	"github.com/Kapeland/task-Astral/internal/storage"
	"github.com/Kapeland/task-Astral/internal/storage/file-storage/file"
	"github.com/Kapeland/task-Astral/internal/storage/file-storage/file_provider"
	"github.com/Kapeland/task-Astral/internal/storage/repository/postgresql/auth"
	"github.com/Kapeland/task-Astral/internal/storage/repository/postgresql/files"
	"github.com/Kapeland/task-Astral/internal/storage/repository/postgresql/users"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
)

func Start() error {
	ctx := context.Background()
	dbStor, err := storage.NewDbStorage(ctx)
	if err != nil {
		logger.Log(logger.ErrPrefix, fmt.Sprintf("App: Start: NewDbStorage: %s", err.Error()))
		return err
	}
	defer dbStor.Close(ctx)

	filesRepo := files.New(dbStor.DB)
	usersRepo := users.New(dbStor.DB)
	authRepo := auth.New(dbStor.DB)

	f := file_provider.NewFileProvider()

	fr := file.NewRepository(f)

	fileStorage := storage.NewFileStorage(filesRepo, fr)
	authStorage := storage.NewAuthStorage(authRepo)
	usersStorage := storage.NewUsersStorage(usersRepo)

	fmdl := models.NewModelFiles(&fileStorage, &usersStorage, &authStorage)
	amdl := models.NewModelAuth(&authStorage, &usersStorage)
	umdl := models.NewModelUsers(&usersStorage)

	serv := services.NewService(&fmdl, &amdl, &umdl)

	serv.Launch()

	return nil
}
