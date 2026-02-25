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

func (app *Application) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// start scheduler in background
	go scheduler.TaskAssignment(ctx)

	rootRouter := chi.NewRouter()

	db, err := dbConfig.SetupDB()
	if err != nil {
		fmt.Println("Error setting up database.")
		return err
	}

	for _, registerFn := range router.DomainRegistries {
		registerFn(db, rootRouter)
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
	// return server.ListenAndServe()
	return err
}
