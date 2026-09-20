package app

import (
	"context"
	"go_project_structure/common_pkg/logger"
	config "go_project_structure/config/env"
	"go_project_structure/di"
	"os/signal"
	"syscall"
)

// global declaration
var applicationLog = logger.Log.Scope("", "app", "application")

type Config struct {
	Addr string // PORT
}

// NewConfig builds a Config from environment variables.
func NewConfig() Config {
	port := config.GetString("PORT", ":8080")
	return Config{
		Addr: port,
	}
}

type Application struct {
	Config Config
}

func NewApplication(cfg Config) Application {
	return Application{
		Config: cfg,
	}
}

// starts the HTTP server. It blocks until a SIGINT/SIGTERM signal is received.
func (app *Application) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.InitializeLogger()

	rootRouter, err := di.BuildApplicationModules(ctx)
	if err != nil {
		return err
	}

	return app.RunServer(ctx, rootRouter)
}
