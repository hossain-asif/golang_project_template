package app

import (
	"context"
	"fmt"
	"go_project_structure/common_pkg/scheduler"
	dbConfig "go_project_structure/config/db"
	config "go_project_structure/config/env"
	"go_project_structure/internal/router"
	"os/signal"
	"syscall"

	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

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

func initModuleRegistry(ctx context.Context) (*chi.Mux, error) {
	db, err := dbConfig.SetupDB()
	if err != nil {
		fmt.Println("Error setting up database.")
		return nil, err
	}

	rootRouter := chi.NewRouter()

	var allTasks []scheduler.Task
	for _, m := range router.Modules {
		m.RegisterRoutes(db, rootRouter)
		allTasks = append(allTasks, m.RegisterTasks(db)...)
	}
	go scheduler.TaskAssignment(ctx, allTasks)

	return rootRouter, nil
}

func (app *Application) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rootRouter, err := initModuleRegistry(ctx)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      rootRouter,
		ReadTimeout:  10 * time.Second, // Set read timeout to 10 seconds
		WriteTimeout: 10 * time.Second, // Set write timeout to 10 seconds
	}

	// shutdown server when ctx is cancelled
	go func() {
		<-ctx.Done()
		fmt.Println("shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}()

	fmt.Println("server running on port", app.Config.Addr, "...")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Println("server error:", err)
	}
	fmt.Println("server stopped")
	return err
}
