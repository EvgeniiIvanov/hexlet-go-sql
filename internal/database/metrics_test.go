package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// TestMetrics is a test metrics collector that records calls
type TestMetrics struct {
	calls []MetricsCall
}

type MetricsCall struct {
	Duration time.Duration
	Err      error
	Query    string
}

func (m *TestMetrics) ObserveQuery(duration time.Duration, err error, query string) {
	m.calls = append(m.calls, MetricsCall{
		Duration: duration,
		Err:      err,
		Query:    query,
	})
}

func TestMetricsWrapper_RecordsQueries(t *testing.T) {
	// Setup
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create metrics collector
	metrics := &TestMetrics{}

	// Wrap database with metrics
	wrappedDB := NewMetricsWrapper(db, metrics)

	// Execute some queries
	ctx := context.Background()

	// Test 1: ExecContext
	_, err = wrappedDB.ExecContext(ctx, "INSERT INTO test (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// Test 2: QueryContext
	rows, err := wrappedDB.QueryContext(ctx, "SELECT * FROM test")
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	rows.Close()

	// Test 3: QueryRowContext
	row := wrappedDB.QueryRowContext(ctx, "SELECT name FROM test WHERE id = ?", 1)
	var name string
	row.Scan(&name)

	// Test 4: Ping
	err = wrappedDB.PingContext(ctx)
	if err != nil {
		t.Fatalf("failed to ping: %v", err)
	}

	// Verify metrics were recorded
	if len(metrics.calls) != 4 {
		t.Errorf("expected 4 metrics calls, got %d", len(metrics.calls))
	}

	// Verify first call (INSERT)
	if metrics.calls[0].Query != "INSERT INTO test (name) VALUES (?)" {
		t.Errorf("unexpected query: %s", metrics.calls[0].Query)
	}
	if metrics.calls[0].Err != nil {
		t.Errorf("expected no error, got %v", metrics.calls[0].Err)
	}
	if metrics.calls[0].Duration == 0 {
		t.Error("expected non-zero duration")
	}

	// Verify second call (SELECT)
	if metrics.calls[1].Query != "SELECT * FROM test" {
		t.Errorf("unexpected query: %s", metrics.calls[1].Query)
	}

	// Verify third call (SELECT with WHERE)
	if metrics.calls[2].Query != "SELECT name FROM test WHERE id = ?" {
		t.Errorf("unexpected query: %s", metrics.calls[2].Query)
	}

	// Verify fourth call (PING)
	if metrics.calls[3].Query != "PING" {
		t.Errorf("unexpected query: %s", metrics.calls[3].Query)
	}
}

func TestMetricsWrapper_RecordsErrors(t *testing.T) {
	// Setup
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create metrics collector
	metrics := &TestMetrics{}

	// Wrap database with metrics
	wrappedDB := NewMetricsWrapper(db, metrics)

	// Execute invalid query (should fail)
	ctx := context.Background()
	_, err = wrappedDB.ExecContext(ctx, "INVALID SQL QUERY")

	// Verify error was recorded
	if len(metrics.calls) != 1 {
		t.Errorf("expected 1 metrics call, got %d", len(metrics.calls))
	}

	if metrics.calls[0].Err == nil {
		t.Error("expected error to be recorded, got nil")
	}
}

func TestNoOpMetrics_DoesNotPanic(t *testing.T) {
	// Setup
	metrics := &NoOpMetrics{}

	// This should not panic
	metrics.ObserveQuery(time.Millisecond, nil, "SELECT 1")
	metrics.ObserveQuery(time.Second, sql.ErrNoRows, "SELECT * FROM nowhere")
}
