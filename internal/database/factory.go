package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"example.com/go-sql/internal/db"
	"example.com/go-sql/internal/migrate"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// Connection represents a database connection with its metadata
type Connection struct {
	DB      *sql.DB
	Queries *db.Queries
	Driver  string
}

// NewConnection creates a new database connection based on environment variables
// It supports both SQLite and PostgreSQL:
// - DATABASE_URL starting with "postgres://" => PostgreSQL
// - DATABASE_URL with file path or empty => SQLite
// - DB_PATH => SQLite (legacy)
func NewConnection(ctx context.Context) (*Connection, error) {
	dbURL := os.Getenv("DATABASE_URL")

	// Legacy support: check DB_PATH if DATABASE_URL is not set
	if dbURL == "" {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "./data.db"
		}
		return newSQLiteConnection(ctx, dbPath)
	}

	// Determine database type from URL
	if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
		return newPostgresConnection(ctx, dbURL)
	}

	// Assume SQLite for file paths
	return newSQLiteConnection(ctx, dbURL)
}

// newSQLiteConnection creates a SQLite connection
func newSQLiteConnection(ctx context.Context, dbPath string) (*Connection, error) {
	// Build connection string with pragmas
	connStr := buildSQLiteConnString(dbPath)

	sqlDB, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to open: %w", err)
	}

	// Configure connection pool for SQLite
	sqlDB.SetMaxOpenConns(1) // SQLite works best with single connection
	sqlDB.SetMaxIdleConns(1)

	// Test connection
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("sqlite: failed to ping: %w", err)
	}

	// Run migrations using goose
	if err := migrate.Up(sqlDB, "sqlite"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("sqlite: migrations failed: %w", err)
	}

	// Wrap with metrics
	metricsDB := NewMetricsWrapper(sqlDB, nil) // nil = use NoOpMetrics

	return &Connection{
		DB:      sqlDB,
		Queries: db.New(metricsDB),
		Driver:  "sqlite",
	}, nil
}

// newPostgresConnection creates a PostgreSQL connection
func newPostgresConnection(ctx context.Context, dbURL string) (*Connection, error) {
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to open: %w", err)
	}

	// Configure connection pool for PostgreSQL
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	// Test connection
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("postgres: failed to ping: %w", err)
	}

	// Run migrations using goose
	if err := migrate.Up(sqlDB, "postgres"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("postgres: migrations failed: %w", err)
	}

	// Layer 1: Wrap with metrics
	metricsDB := NewMetricsWrapper(sqlDB, nil) // nil = use NoOpMetrics

	// Layer 2: Wrap with PostgreSQL placeholder converter (on top of metrics)
	// This converts SQLite-style placeholders (?) to PostgreSQL-style ($1, $2, etc.)
	wrappedDB := &postgresDBWrapper{db: metricsDB}

	return &Connection{
		DB:      sqlDB,
		Queries: db.New(wrappedDB),
		Driver:  "postgres",
	}, nil
}

// DBExecutor is an interface that defines the methods we need from database connections
// This allows us to compose wrappers (metrics, placeholder conversion, etc.)
type DBExecutor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// postgresDBWrapper wraps a DBExecutor and translates SQLite placeholders (?) to PostgreSQL placeholders ($1, $2, etc.)
type postgresDBWrapper struct {
	db DBExecutor
}

func (w *postgresDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	converted := convertPlaceholders(query)
	return w.db.QueryContext(ctx, converted, args...)
}

func (w *postgresDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	converted := convertPlaceholders(query)
	return w.db.QueryRowContext(ctx, converted, args...)
}

func (w *postgresDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	converted := convertPlaceholders(query)
	return w.db.ExecContext(ctx, converted, args...)
}

func (w *postgresDBWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	converted := convertPlaceholders(query)
	return w.db.PrepareContext(ctx, converted)
}

// convertPlaceholders converts SQLite-style ? placeholders to PostgreSQL-style $1, $2, etc.
func convertPlaceholders(query string) string {
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

	converted := result.String()
	// Debug: uncomment to see query conversion
	// log.Printf("DEBUG: Converted query:\n%s\nTO:\n%s\n", query, converted)
	return converted
}

// buildSQLiteConnString builds a SQLite connection string with pragmas
func buildSQLiteConnString(dbPath string) string {
	connStr := dbPath
	if !strings.Contains(connStr, "?") {
		connStr += "?"
	}

	// Add foreign key support
	if !strings.Contains(connStr, "_foreign_keys") {
		if !strings.HasSuffix(connStr, "?") {
			connStr += "&"
		}
		connStr += "_foreign_keys=on"
	}

	// Add busy timeout
	if !strings.Contains(connStr, "_busy_timeout") {
		connStr += "&_busy_timeout=5000"
	}

	return connStr
}
