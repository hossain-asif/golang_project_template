package app

import (
	"context"
	"go_project_structure/common_pkg/logger"
	"go_project_structure/common_pkg/scheduler"
	"go_project_structure/common_pkg/storage"
	dbConfig "go_project_structure/config/database"
	config "go_project_structure/config/env"
	"go_project_structure/internal/router"
	"os/signal"
	"syscall"

	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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

func initModuleRegistry(ctx context.Context, fs *storage.FileStore) (*chi.Mux, error) {

	// setup mongodb
	// hook, err := dbConfig.SetupMongoDB()
	// if err != nil {
	// 	return nil, err
	// }
	// logger.AddHook(hook)

	// set up postgres db
	db, err := dbConfig.SetupDB()
	if err != nil {

		// logger.Log.WithFields(map[string]interface{}{
		// 	// "layer":     "",
		// 	"module":    "app",
		// 	"component": "application",
		// 	"method":    "initModuleRegistry",
		// 	"error":     err,
		// }).Error("Error setting up database.")
		appLog.Method("initModuleRegistry").WithError(err).Error("Error setting up database.")

		return nil, err
	}

	rootRouter := chi.NewRouter()

	var allTasks []scheduler.Task
	for _, m := range router.Modules {
		m.RegisterRoutes(db, rootRouter, fs)
		allTasks = append(allTasks, m.RegisterTasks(db, fs)...)
	}
	go scheduler.TaskAssignment(ctx, allTasks)

	return rootRouter, nil
}

func (app *Application) Run() error {
	// shutdown server on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// logger setup
	logger.InitLogger()

	// Build the index at startup (scans file ONCE)
	fileDirectory := config.GetString("KV_FILE_DIRECTORY", "")
	fs, err := storage.NewFileStore(fileDirectory, storage.FormatKV)
	if err != nil {
		appLog.Errorf("Failed to build file index: %v", err)
		return err
	}
	defer fs.Close()

	// Register as default so controllers can use file_system.GetRecord()
	storage.SetDefault(fs)

	// initialize module registry
	rootRouter, err := initModuleRegistry(ctx, fs)
	if err != nil {
		return err
	}

	// server configuration
	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      rootRouter,
		ReadTimeout:  10 * time.Second, // Set read timeout to 10 seconds
		WriteTimeout: 10 * time.Second, // Set write timeout to 10 seconds
	}

	// shutdown server when ctx is cancelled
	go func() {
		<-ctx.Done()

		// logger.Log.WithFields(map[string]interface{}{
		// 	// "layer":     "",
		// 	"module":    "app",
		// 	"component": "application",
		// 	"method":    "Run",
		// }).Warn("shutting down server...")
		appLog.Method("Run").Warn("shutting down server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()

	// logger.Log.WithFields(map[string]interface{}{
	// 	// "layer":     "",
	// 	"module":    "app",
	// 	"component": "application",
	// 	"method":    "Run",
	// 	"port":      app.Config.Addr,
	// }).Info("server running on given port.")
	appLog.Method("Run").Infof("server running on port: %s", app.Config.Addr)

	if serverErr := server.ListenAndServe(); serverErr != http.ErrServerClosed {
		// logger.Log.WithFields(map[string]interface{}{
		// 	// "layer":     "",
		// 	"module":    "app",
		// 	"component": "application",
		// 	"method":    "Run",
		// 	"error":     serverErr,
		// }).Error("server initialization failed")

		appLog.Method("Run").WithError(serverErr).Error("server initialization failed")

		return serverErr
	}

	// logger.Log.WithFields(map[string]interface{}{
	// 	// "layer":     "",
	// 	"module":    "app",
	// 	"component": "application",
	// 	"method":    "Run",
	// }).Info("server stopped.")

	appLog.Method("Run").Info("server stopped.")

	return err
}
