//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// SetupIntegrationDB creates a real SQLite database file for integration tests
// Returns the database connection and cleanup function
func SetupIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx := context.Background()

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration_test.db")

	// Open real database file (not in-memory)
	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open integration test database: %v", err)
	}

	// Run migrations
	if err := runIntegrationMigrations(ctx, db); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Clean up on test completion
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// runIntegrationMigrations executes all migration files for integration tests
func runIntegrationMigrations(ctx context.Context, db *sql.DB) error {
	// Try different relative paths to find migrations
	// Tests might run from different directories
	possiblePaths := [][]string{
		{"migrations/001_schema.sql", "migrations/002_add_orders.sql"},
		{"../../migrations/001_schema.sql", "../../migrations/002_add_orders.sql"},
		{"../../../migrations/001_schema.sql", "../../../migrations/002_add_orders.sql"},
	}

	var migrations []string
	for _, paths := range possiblePaths {
		if _, err := os.Stat(paths[0]); err == nil {
			migrations = paths
			break
		}
	}

	if migrations == nil {
		return fmt.Errorf("could not find migrations directory")
	}

	for _, migration := range migrations {
		data, err := os.ReadFile(migration)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", migration, err)
		}

		if _, err := db.ExecContext(ctx, string(data)); err != nil {
			return fmt.Errorf("execute migration %s: %w", migration, err)
		}
	}

	return nil
}

// WithTx runs a test function inside a transaction and rolls it back
// This provides test isolation without polluting the database
func WithTx(t *testing.T, db *sql.DB, fn func(tx *sql.Tx)) {
	t.Helper()

	ctx := context.Background()

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Ensure rollback happens
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Errorf("failed to rollback transaction: %v", err)
		}
	}()

	// Run test function
	fn(tx)
}

// WithTxContext runs a test function inside a transaction with context
// Use this when you need to pass context to repository methods
func WithTxContext(t *testing.T, db *sql.DB, fn func(ctx context.Context, tx *sql.Tx)) {
	t.Helper()

	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Errorf("failed to rollback transaction: %v", err)
		}
	}()

	fn(ctx, tx)
}

// IntegrationTestData holds test fixtures for integration tests
type IntegrationTestData struct {
	Users   []IntegrationUser
	Courses []IntegrationCourse
}

type IntegrationUser struct {
	Email string
	Name  string
	Age   int64
}

type IntegrationCourse struct {
	Slug  string
	Title string
	Price int64
}

// GetDefaultTestData returns realistic test data for integration tests
func GetDefaultTestData() IntegrationTestData {
	return IntegrationTestData{
		Users: []IntegrationUser{
			{Email: "alice@example.com", Name: "Alice Smith", Age: 25},
			{Email: "bob@example.com", Name: "Bob Johnson", Age: 30},
			{Email: "charlie@example.com", Name: "Charlie Brown", Age: 22},
		},
		Courses: []IntegrationCourse{
			{Slug: "go-fundamentals", Title: "Go Programming Fundamentals", Price: 9999},
			{Slug: "advanced-go", Title: "Advanced Go Patterns", Price: 14999},
			{Slug: "go-microservices", Title: "Building Microservices with Go", Price: 19999},
		},
	}
}
