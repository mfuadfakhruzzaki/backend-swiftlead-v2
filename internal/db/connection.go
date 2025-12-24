package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// Connect establishes a connection to the database
func Connect(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to database successfully")

	// Enable TimescaleDB extension
	if err := enableTimescaleDB(db); err != nil {
		logger.Warn("TimescaleDB extension not available: %v", err)
	}

	return db, nil
}

// enableTimescaleDB enables the TimescaleDB extension
func enableTimescaleDB(db *sql.DB) error {
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE")
	if err != nil {
		return err
	}
	logger.Info("TimescaleDB extension enabled")
	return nil
}

// Close closes the database connection
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
