package app

import (
	"context"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"
	config "go_project_structure/config/env"
	"go_project_structure/internal/module"
	"os/signal"
	"syscall"
)

// global declaration
// better to keep the log name as similar to file name
var appLog = logger.Log.Scope("", "app", "application")

// Config holds the configuration for the server.
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

// Application is the top-level wiring object.
// Modules is the ordered list of domain modules supplied by the composition root (main.go).
type Application struct {
	Config  Config
	Modules []module.Module
}

// NewApplication constructs an Application with its config and module list.
func NewApplication(cfg Config, modules []module.Module) Application {
	return Application{
		Config:  cfg,
		Modules: modules,
	}
}

// Run initialises all infrastructure, bootstraps modules, and starts the HTTP
// server. It blocks until a SIGINT/SIGTERM signal is received.
func (app *Application) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.InitLogger()

	// Connect MongoDB as a logrus hook (uncomment when needed)
	// mongoHook, err := setupMongoHook()
	// if err != nil {
	// 	return err
	// }
	// defer mongoHook.Disconnect()

	dep, cleanup, err := app.setupInfrastructure()
	if err != nil {
		return err
	}
	defer cleanup()

	rootRouter, allTasks, err := BootstrapModules(app.Modules, dep)
	if err != nil {
		return err
	}

	go scheduler.TaskAssignment(ctx, allTasks)

	return app.RunServer(ctx, rootRouter)
}

// setupInfrastructure initialises shared infra (file store, database) and
// returns a populated Dependency bundle together with a cleanup function.
func (app *Application) setupInfrastructure() (module.Dependency, func(), error) {
	fs, err := SetupFileStore()
	if err != nil {
		return module.Dependency{}, nil, err
	}
	storage.SetDefault(fs)

	db, err := SetupDB()
	if err != nil {
		fs.Close()
		return module.Dependency{}, nil, err
	}

	dep := module.Dependency{DB: db, FS: fs}
	cleanup := func() { fs.Close() }
	return dep, cleanup, nil
}
