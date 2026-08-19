package app

import (
	"context"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/common_pkg/scheduler"
	config "go_project_structure/config/env"
	"go_project_structure/internal/module"
	"os/signal"
	"syscall"
)

// global declaration
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

	dep, err := app.setupInfrastructure()
	if err != nil {
		return err
	}

	rootRouter, allTasks, err := dependencyInit(app.Modules, dep)
	if err != nil {
		return err
	}

	go scheduler.TaskAssignment(ctx, allTasks)

	return app.RunServer(ctx, rootRouter)
}

// setupInfrastructure initialises shared infra (file store, database) and
// returns a populated Dependency bundle together with a cleanup function.
func (app *Application) setupInfrastructure() (module.Dependency, error) {

	db, err := SetupDB()
	if err != nil {
		return module.Dependency{}, err
	}



	// Connect MongoDB as a logrus hook (uncomment when needed)
	// mongoHook, err := setupMongoHook()
	// if err != nil {
	// 	return err
	// }
	// defer mongoHook.Disconnect()


	redisClient, err := SetupRedis()
	if err != nil {
		return module.Dependency{}, err
	}

	dep := module.Dependency{
		DB: db,
		RedisClient: redisClient,
	}
	return dep, nil
}
