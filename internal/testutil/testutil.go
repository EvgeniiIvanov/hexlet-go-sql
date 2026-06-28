package testutil

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"example.com/go-sql/internal/migrate"

	_ "modernc.org/sqlite"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Use in-memory database for tests
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Run migrations using goose
	if err := migrate.Up(db, "sqlite"); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Clean up on test completion
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TruncateTable removes all data from a table (for test isolation)
func TruncateTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()

	_, err := db.Exec("DELETE FROM " + table)
	if err != nil {
		t.Fatalf("failed to truncate table %s: %v", table, err)
	}
}

// SetupTestDBWithData creates a test database and seeds it with test data
func SetupTestDBWithData(t *testing.T) (*sql.DB, TestData) {
	t.Helper()

	db := SetupTestDB(t)
	data := SeedTestData(t, db)

	return db, data
}

// TestData holds references to seeded test data
type TestData struct {
	UserID1   int64
	UserID2   int64
	CourseID1 int64
	CourseID2 int64
}

// SeedTestData inserts common test data
func SeedTestData(t *testing.T, db *sql.DB) TestData {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := TestData{}

	// Insert test users
	result, err := db.ExecContext(ctx,
		"INSERT INTO users (email, name, age) VALUES (?, ?, ?)",
		"alice@test.com", "Alice", 25)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	data.UserID1, _ = result.LastInsertId()

	result, err = db.ExecContext(ctx,
		"INSERT INTO users (email, name, age) VALUES (?, ?, ?)",
		"bob@test.com", "Bob", 30)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	data.UserID2, _ = result.LastInsertId()

	// Insert test courses
	result, err = db.ExecContext(ctx,
		"INSERT INTO courses (slug, title, price) VALUES (?, ?, ?)",
		"go-101", "Go Programming", 9999)
	if err != nil {
		t.Fatalf("failed to insert test course: %v", err)
	}
	data.CourseID1, _ = result.LastInsertId()

	result, err = db.ExecContext(ctx,
		"INSERT INTO courses (slug, title, price) VALUES (?, ?, ?)",
		"rust-101", "Rust Programming", 14999)
	if err != nil {
		t.Fatalf("failed to insert test course: %v", err)
	}
	data.CourseID2, _ = result.LastInsertId()

	return data
}
