package app

import (
	"context"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"
	config "go_project_structure/config/env"
	"go_project_structure/internal/router"
	"os/signal"
	"syscall"
)

// global decalaration
// better to keep the log name as similar to file name
var appLog = logger.Log.Scope("", "app", "application")

// Config holds the configuration for the server.
type Config struct {
	Addr string // PORT
}

// constructor for Config
func NewConfig() Config {
	port := config.GetString("PORT", ":8080")
	return Config{
		Addr: port,
	}
}

type Application struct {
	Config Config
}

// constructor for Application
func NewApplication(config Config) Application {
	return Application{
		Config: config,
	}
}

func (app *Application) Run() error {
	// shutdown server on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// logger setup
	logger.InitLogger()

	// Connect MongoDB as a logrus hook
	// mongoHook, err := setupMongoHook()
	// if err != nil {
	// 	return err
	// }
	// defer mongoHook.Disconnect()

	// Init file store
	fs, err := SetupFileStore()
	if err != nil {
		return err
	}
	defer fs.Close()
	storage.SetDefault(fs)

	// Init database
	db, err := SetupDB()
	if err != nil {
		return err
	}

	// Bootstrap modules → build root router + collect tasks
	rootRouter, allTasks, err := BootstrapModules(router.Modules, db, fs)
	if err != nil {
		return err
	}

	// Start background task scheduler
	go scheduler.TaskAssignment(ctx, allTasks)

	return app.RunServer(ctx, rootRouter)
}
