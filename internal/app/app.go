package app

import (
	"context"
	"flag"
	"github.com/Kapeland/task-Astral/internal/models"
	"github.com/Kapeland/task-Astral/internal/services"
	"github.com/Kapeland/task-Astral/internal/storage"
	"github.com/Kapeland/task-Astral/internal/storage/file-storage/file"
	"github.com/Kapeland/task-Astral/internal/storage/file-storage/file_provider"
	"github.com/Kapeland/task-Astral/internal/storage/repository/postgresql/auth"
	"github.com/Kapeland/task-Astral/internal/storage/repository/postgresql/files"
	"github.com/Kapeland/task-Astral/internal/storage/repository/postgresql/users"
	"github.com/Kapeland/task-Astral/internal/utils/config"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"github.com/pressly/goose/v3"
)

func Start(cfg *config.Config, lgr *logger.Logger) error {
	migration := flag.Bool("migration", true, "Defines the migration start option")
	flag.Parse()

	ctx := context.Background()
	dbStor, err := storage.NewPostgresStorage(ctx)
	if err != nil {
		lgr.Error(err.Error(), "App", "Start", "NewPostgresStorage")
		return err
	}
	defer dbStor.Close()
	if *migration {
		if err := goose.Up(dbStor.DB.GetDB().DB, cfg.Database.Migrations); err != nil {
			lgr.Error("Migration failed: "+err.Error(), "App", "Start", " goose.Up")

			return err
		}
	}

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

	serv.Launch(cfg, lgr)

	return nil
}
