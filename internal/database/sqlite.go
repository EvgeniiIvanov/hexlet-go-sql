package database

import (
	"context"
	"database/sql"
	"fmt"
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

// InitSchema initializes the database schema
func InitSchema(ctx context.Context, db *sql.DB) error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			price INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			name TEXT,
			age INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.ExecContext(ctx, schema); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	return nil
}
