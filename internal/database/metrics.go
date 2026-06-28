// Package database provides database connection management with pluggable metrics.
//
// The MetricsWrapper allows you to collect metrics for all database operations
// without modifying your application code. By default, it uses NoOpMetrics (no overhead),
// but you can easily swap in Prometheus, StatsD, or any custom metrics implementation.
//
// Architecture:
//
//	Application Code
//	       ↓
//	  sqlc Queries (internal/db)
//	       ↓
//	  MetricsWrapper (this package) - Measures query duration
//	       ↓
//	  PostgresWrapper (if PostgreSQL) - Converts ? to $1, $2, ...
//	       ↓
//	  sql.DB (database/sql)
//	       ↓
//	  Database (SQLite or PostgreSQL)
//
// Usage:
//
//	// Default (no metrics)
//	conn, _ := database.NewConnection(ctx)
//
//	// With custom metrics (requires modifying factory.go)
//	metricsDB := database.NewMetricsWrapper(sqlDB, &MyMetrics{})
package database

import (
	"context"
	"database/sql"
	"time"
)

// MetricsCollector is an interface for collecting database metrics
// This allows easy swapping of metrics implementations (Prometheus, StatsD, etc.)
//
// Implement this interface to add custom metrics collection:
//   - duration: how long the query took
//   - err: error if query failed, nil if successful
//   - query: the SQL query string (useful for categorization)
type MetricsCollector interface {
	ObserveQuery(duration time.Duration, err error, query string)
}

// NoOpMetrics is a no-op metrics collector (default behavior)
type NoOpMetrics struct{}

func (m *NoOpMetrics) ObserveQuery(duration time.Duration, err error, query string) {
	// No-op: do nothing
	// This can be replaced with actual metrics implementation later
}

// Example: LogMetrics is a simple logging metrics collector (useful for debugging)
// To use: pass &LogMetrics{} instead of nil when creating NewMetricsWrapper
//
// Example implementation of a real metrics collector:
//
// import (
//     "log"
//     "strings"
//     "time"
// )
//
// type LogMetrics struct{}
//
// func (m *LogMetrics) ObserveQuery(duration time.Duration, err error, query string) {
//     queryType := extractQueryType(query)
//     status := "success"
//     if err != nil {
//         status = "error"
//     }
//     log.Printf("[METRICS] query=%s duration=%v status=%s", queryType, duration, status)
// }
//
// func extractQueryType(query string) string {
//     query = strings.TrimSpace(query)
//     if len(query) > 20 {
//         query = query[:20] + "..."
//     }
//     parts := strings.Fields(query)
//     if len(parts) > 0 {
//         return strings.ToUpper(parts[0])
//     }
//     return "UNKNOWN"
// }
//
// Example with Prometheus:
//
// import (
//     "github.com/prometheus/client_golang/prometheus"
//     "github.com/prometheus/client_golang/prometheus/promauto"
// )
//
// var (
//     dbQueryDuration = promauto.NewHistogramVec(
//         prometheus.HistogramOpts{
//             Name: "db_query_duration_seconds",
//             Help: "Database query duration in seconds",
//         },
//         []string{"query_type", "status"},
//     )
// )
//
// type PrometheusMetrics struct{}
//
// func (m *PrometheusMetrics) ObserveQuery(duration time.Duration, err error, query string) {
//     queryType := extractQueryType(query)
//     status := "success"
//     if err != nil {
//         status = "error"
//     }
//     dbQueryDuration.WithLabelValues(queryType, status).Observe(duration.Seconds())
// }

// MetricsWrapper wraps sql.DB and collects metrics for all database operations
type MetricsWrapper struct {
	db      *sql.DB
	metrics MetricsCollector
}

// NewMetricsWrapper creates a new metrics wrapper around a database connection
func NewMetricsWrapper(db *sql.DB, metrics MetricsCollector) *MetricsWrapper {
	if metrics == nil {
		metrics = &NoOpMetrics{}
	}
	return &MetricsWrapper{
		db:      db,
		metrics: metrics,
	}
}

// ExecContext wraps ExecContext with metrics
func (m *MetricsWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := m.db.ExecContext(ctx, query, args...)
	m.metrics.ObserveQuery(time.Since(start), err, query)
	return res, err
}

// QueryContext wraps QueryContext with metrics
func (m *MetricsWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := m.db.QueryContext(ctx, query, args...)
	m.metrics.ObserveQuery(time.Since(start), err, query)
	return rows, err
}

// QueryRowContext wraps QueryRowContext with metrics
func (m *MetricsWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := m.db.QueryRowContext(ctx, query, args...)
	// Note: QueryRowContext doesn't return an error immediately, error is deferred to Scan
	// We still record the metrics, but err will always be nil here
	m.metrics.ObserveQuery(time.Since(start), nil, query)
	return row
}

// PrepareContext wraps PrepareContext with metrics
func (m *MetricsWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	start := time.Now()
	stmt, err := m.db.PrepareContext(ctx, query)
	m.metrics.ObserveQuery(time.Since(start), err, query)
	return stmt, err
}

// BeginTx wraps BeginTx and returns a metrics-wrapped transaction
func (m *MetricsWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	start := time.Now()
	tx, err := m.db.BeginTx(ctx, opts)
	m.metrics.ObserveQuery(time.Since(start), err, "BEGIN")
	return tx, err
}

// Ping wraps Ping with metrics
func (m *MetricsWrapper) Ping() error {
	start := time.Now()
	err := m.db.Ping()
	m.metrics.ObserveQuery(time.Since(start), err, "PING")
	return err
}

// PingContext wraps PingContext with metrics
func (m *MetricsWrapper) PingContext(ctx context.Context) error {
	start := time.Now()
	err := m.db.PingContext(ctx)
	m.metrics.ObserveQuery(time.Since(start), err, "PING")
	return err
}

// Close wraps Close
func (m *MetricsWrapper) Close() error {
	return m.db.Close()
}

// SetMaxOpenConns wraps SetMaxOpenConns
func (m *MetricsWrapper) SetMaxOpenConns(n int) {
	m.db.SetMaxOpenConns(n)
}

// SetMaxIdleConns wraps SetMaxIdleConns
func (m *MetricsWrapper) SetMaxIdleConns(n int) {
	m.db.SetMaxIdleConns(n)
}

// SetConnMaxLifetime wraps SetConnMaxLifetime
func (m *MetricsWrapper) SetConnMaxLifetime(d time.Duration) {
	m.db.SetConnMaxLifetime(d)
}

// SetConnMaxIdleTime wraps SetConnMaxIdleTime
func (m *MetricsWrapper) SetConnMaxIdleTime(d time.Duration) {
	m.db.SetConnMaxIdleTime(d)
}

// Stats wraps Stats
func (m *MetricsWrapper) Stats() sql.DBStats {
	return m.db.Stats()
}

// DB returns the underlying sql.DB for cases where you need direct access
func (m *MetricsWrapper) DB() *sql.DB {
	return m.db
}
