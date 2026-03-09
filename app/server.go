package app

import (
	"context"
	"net/http"
	"time"
)

// runServer starts the HTTP server and blocks until ctx is cancelled.
func (app *Application) RunServer(ctx context.Context, handler http.Handler) error {
	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown goroutine
	go func() {
		<-ctx.Done()
		appLog.Method("runServer").Warn("shutdown signal received, stopping server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			appLog.Method("runServer").WithError(err).Error("graceful shutdown failed.")
		}
	}()

	appLog.Method("runServer").Infof("server running on port: %s", app.Config.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		appLog.Method("runServer").WithError(err).Error("server initialization failed.")
		return err
	}

	appLog.Method("runServer").Info("server stopped.")
	return nil
}
