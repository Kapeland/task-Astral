package main

import (
	"github.com/Kapeland/task-Astral/internal/app"
	"github.com/Kapeland/task-Astral/internal/utils/config"
	"github.com/Kapeland/task-Astral/internal/utils/logger"
	"log"
	"log/slog"
)

func main() {
	if err := config.ReadConfigYAML(); err != nil {
		log.Fatal("Failed init configuration")
	}
	cfg := config.GetConfig()
	lgr := logger.CreateLogger(&cfg)

	lgr.Info("app started", "main", "", "")
	err := app.Start(&cfg, &lgr)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	lgr.Info("app finished", "main", "", "")
}
