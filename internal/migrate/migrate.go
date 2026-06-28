// Package migrate provides database migration functionality using goose.
//
// This package wraps goose to provide a simple interface for running migrations
// against both SQLite and PostgreSQL databases.
//
// Usage:
//
//	db, _ := sql.Open("sqlite", "app.db")
//	if err := migrate.Up(db, "sqlite"); err != nil {
//	    log.Fatal(err)
//	}
package migrate

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var sqliteMigrations embed.FS

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

// Up runs all available migrations for the specified database driver
func Up(db *sql.DB, driver string) error {
	var fs embed.FS
	var dir string

	switch driver {
	case "sqlite":
		fs = sqliteMigrations
		dir = "migrations"
		goose.SetDialect("sqlite3")
	case "postgres", "postgresql":
		fs = postgresMigrations
		dir = "migrations/postgres"
		goose.SetDialect("postgres")
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	goose.SetBaseFS(fs)

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Down rolls back the most recent migration
func Down(db *sql.DB, driver string) error {
	var fs embed.FS
	var dir string

	switch driver {
	case "sqlite":
		fs = sqliteMigrations
		dir = "migrations"
		goose.SetDialect("sqlite3")
	case "postgres", "postgresql":
		fs = postgresMigrations
		dir = "migrations/postgres"
		goose.SetDialect("postgres")
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	goose.SetBaseFS(fs)

	if err := goose.Down(db, dir); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// Status prints the migration status
func Status(db *sql.DB, driver string) error {
	var fs embed.FS
	var dir string

	switch driver {
	case "sqlite":
		fs = sqliteMigrations
		dir = "migrations"
		goose.SetDialect("sqlite3")
	case "postgres", "postgresql":
		fs = postgresMigrations
		dir = "migrations/postgres"
		goose.SetDialect("postgres")
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	goose.SetBaseFS(fs)

	if err := goose.Status(db, dir); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// Version returns the current migration version
func Version(db *sql.DB, driver string) (int64, error) {
	switch driver {
	case "sqlite":
		goose.SetDialect("sqlite3")
	case "postgres", "postgresql":
		goose.SetDialect("postgres")
	default:
		return 0, fmt.Errorf("unsupported database driver: %s", driver)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, fmt.Errorf("failed to get database version: %w", err)
	}

	return version, nil
}
