package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/swiftlead/backend-swiftlet/internal/app"
	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/db"
	"github.com/swiftlead/backend-swiftlet/internal/router"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New(cfg.LogLevel)
	logger.SetDefault(log)

	log.Info("Starting Swiftlet Backend...")
	log.Info("Environment: %s", cfg.Env)

	// Connect to database
	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close(database)

	// Run migrations
	log.Info("Running database migrations...")
	if err := db.Migrate(database); err != nil {
		log.Fatal("Failed to run migrations: %v", err)
	}

	// Create application container
	container := app.NewContainer(cfg, database)
	defer container.Close()

	// Initialize optional services
	if err := container.InitStorage(); err != nil {
		log.Warn("MinIO storage not available: %v", err)
	} else {
		log.Info("MinIO storage initialized")
	}

	if err := container.ConnectMQTT(); err != nil {
		log.Warn("MQTT broker not available: %v", err)
	} else {
		log.Info("MQTT broker connected")
	}

	// Setup router
	r := router.New(cfg)
	router.SetupRoutes(r, cfg, container)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info("Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
