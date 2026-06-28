//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/go-sql/internal/db"
	"example.com/go-sql/internal/migrate"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// SetupIntegrationDB creates a database for integration tests
// Supports both SQLite and PostgreSQL based on TEST_DATABASE_URL env var
// Returns the database connection and cleanup function
func SetupIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if PostgreSQL test database is configured
	testDBURL := os.Getenv("TEST_DATABASE_URL")

	if testDBURL != "" && (strings.HasPrefix(testDBURL, "postgres://") || strings.HasPrefix(testDBURL, "postgresql://")) {
		return setupPostgresIntegrationDB(t, ctx, testDBURL)
	}

	// Default to SQLite
	return setupSQLiteIntegrationDB(t, ctx)
}

// setupSQLiteIntegrationDB creates a SQLite database for integration tests
func setupSQLiteIntegrationDB(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration_test.db")

	// Open real database file (not in-memory)
	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open SQLite integration test database: %v", err)
	}

	// Run migrations using goose
	if err := migrate.Up(db, "sqlite"); err != nil {
		db.Close()
		t.Fatalf("failed to run SQLite migrations: %v", err)
	}

	// Clean up on test completion
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// setupPostgresIntegrationDB creates a PostgreSQL database for integration tests
func setupPostgresIntegrationDB(t *testing.T, ctx context.Context, dbURL string) *sql.DB {
	t.Helper()

	// Open PostgreSQL connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to open PostgreSQL integration test database: %v", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to ping PostgreSQL: %v (is PostgreSQL running? try: make docker-up)", err)
	}

	// Clean up database before running tests
	cleanupPostgresDB(t, ctx, db)

	// Run migrations using goose
	if err := migrate.Up(db, "postgres"); err != nil {
		db.Close()
		t.Fatalf("failed to run PostgreSQL migrations: %v", err)
	}

	// Clean up on test completion
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		cleanupPostgresDB(t, cleanupCtx, db)
		db.Close()
	})

	return db
}

// cleanupPostgresDB removes all data from PostgreSQL test database
func cleanupPostgresDB(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	// Drop all tables (in correct order due to foreign keys)
	tables := []string{
		"course_reviews",
		"lessons",
		"order_items",
		"enrollments",
		"orders",
		"courses",
		"users",
		"goose_db_version", // goose version tracking table
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			t.Logf("warning: failed to drop table %s: %v", table, err)
		}
	}
}

// WithTx runs a test function inside a transaction and rolls it back
// This provides test isolation without polluting the database
func WithTx(t *testing.T, db *sql.DB, fn func(tx *sql.Tx)) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
func WithTxContext(t *testing.T, sqlDB *sql.DB, fn func(ctx context.Context, tx *sql.Tx)) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := sqlDB.BeginTx(ctx, nil)
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

// postgresTxWrapper wraps sql.Tx and translates SQLite placeholders (?) to PostgreSQL placeholders ($1, $2, etc.)
type postgresTxWrapper struct {
	tx *sql.Tx
}

func (w *postgresTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.tx.QueryContext(ctx, convertPlaceholdersTest(query), args...)
}

func (w *postgresTxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.tx.QueryRowContext(ctx, convertPlaceholdersTest(query), args...)
}

func (w *postgresTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.tx.ExecContext(ctx, convertPlaceholdersTest(query), args...)
}

func (w *postgresTxWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.tx.PrepareContext(ctx, convertPlaceholdersTest(query))
}

// convertPlaceholdersTest converts SQLite-style ? placeholders to PostgreSQL-style $1, $2, etc.
func convertPlaceholdersTest(query string) string {
	var result strings.Builder
	paramIndex := 1
	inString := false
	var stringChar byte

	for i := 0; i < len(query); i++ {
		ch := query[i]

		// Track if we're inside a string literal
		if ch == '\'' || ch == '"' {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				// Check if it's escaped
				if i > 0 && query[i-1] != '\\' {
					inString = false
				}
			}
		}

		// Only replace ? if we're not inside a string
		if ch == '?' && !inString {
			result.WriteString(fmt.Sprintf("$%d", paramIndex))
			paramIndex++
		} else {
			result.WriteByte(ch)
		}
	}

	return result.String()
}

// NewQueries creates a new db.Queries instance that works with both SQLite and PostgreSQL
// It wraps transactions appropriately for PostgreSQL placeholder conversion
func NewQueries(tx *sql.Tx) *db.Queries {
	// Check if we're using PostgreSQL
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL != "" && (strings.HasPrefix(testDBURL, "postgres://") || strings.HasPrefix(testDBURL, "postgresql://")) {
		// Wrap for PostgreSQL
		wrappedTx := &postgresTxWrapper{tx: tx}
		return db.New(wrappedTx)
	}

	// Use raw transaction for SQLite
	return db.New(tx)
}

// GetDriverName returns the current test driver name
func GetDriverName() string {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL != "" && (strings.HasPrefix(testDBURL, "postgres://") || strings.HasPrefix(testDBURL, "postgresql://")) {
		return "postgres"
	}
	return "sqlite"
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
