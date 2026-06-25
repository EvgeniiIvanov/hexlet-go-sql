package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// Config holds database configuration
type Config struct {
	Path            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns default database configuration
func DefaultConfig() Config {
	return Config{
		Path:            "data.db",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxIdleTime: 30 * time.Second,
	}
}

// Connect opens a connection to the SQLite database
func Connect(ctx context.Context, cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", cfg.Path))
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}

// InitSchema initializes the database schema by running migrations
func InitSchema(ctx context.Context, db *sql.DB) error {
	// Read the migration file
	data, err := os.ReadFile("migrations/001_schema.sql")
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	// Execute the migration
	if _, err := db.ExecContext(ctx, string(data)); err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	return nil
}
