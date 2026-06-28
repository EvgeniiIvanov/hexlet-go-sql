# Development Guide

> **TL;DR**: Use `make build` to compile, `make test` for unit tests, `make test-integration` for integration tests. Modify SQL files in `query/` and run `sqlc generate`. Use Docker for PostgreSQL development.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Database Development](#database-development)
- [Docker & PostgreSQL](#docker--postgresql)
- [Code Generation](#code-generation)
- [Metrics & Observability](#metrics--observability)
- [Common Tasks](#common-tasks)
- [Troubleshooting](#troubleshooting)

## Getting Started

> **TL;DR**: Install Go 1.25+, sqlc v1.31.1, goose v3.27.1, clone repo, run `make build`.

### Prerequisites

```bash
# Required
- Go 1.25.0 or higher
- sqlc v1.31.1 (SQL to Go code generator)
- goose v3.27.1 (Database migration tool)

# Optional (for PostgreSQL development)
- Docker & Docker Compose
```

### Installing Tools

```bash
# Install sqlc (standard Go way)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

# Install goose (standard Go way)
go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

# Verify installations
sqlc version
goose -version
```

### Initial Setup

```bash
# Clone repository
git clone <repo-url>
cd hexlet-go-sql

# Install Go dependencies
go mod download

# Generate database code
sqlc generate

# Build
make build

# Run tests
make test

# You're ready! 🎉
```

## Development Workflow

> **TL;DR**: Edit code → Test → Commit. For SQL changes: Edit SQL → `sqlc generate` → Test → Commit.

### Standard Workflow

```bash
# 1. Create a branch
git checkout -b feature/new-feature

# 2. Make changes
# Edit files...

# 3. Run tests
make test              # Fast unit tests
make test-integration  # Integration tests

# 4. Build
make build

# 5. Test manually
./bin/gosql user-list

# 6. Commit
git add .
git commit -m "feat: add new feature"
git push
```

### SQL Development Workflow

When you need to change database queries or schema:

```bash
# 1. Edit SQL files
vim query/users/users_read.sql

# 2. Regenerate Go code
sqlc generate

# 3. Update repository/service if needed
vim internal/repository/repository.go

# 4. Run tests
make test

# 5. Commit both SQL and generated Go files
git add query/ internal/db/
git commit -m "feat: add new user query"
```

### Adding a New Feature

Example: Adding a "get user by email" feature

```bash
# 1. Add SQL query
cat >> query/users/users_read.sql << 'EOF'
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;
EOF

# 2. Generate code
sqlc generate

# 3. Add repository method (if needed - optional, sqlc generated it)
# The method is auto-generated in internal/db/users_read.sql.go

# 4. Add service method
vim internal/service/service.go
# Add:
# func (s *Service) GetUserByEmail(ctx context.Context, email string) (models.User, error)

# 5. Add CLI command
vim cmd/app/main.go
# Add UserGetByEmailCmd struct and Run method

# 6. Test
make test

# 7. Test manually
make build
./bin/gosql user-get-by-email -e "test@example.com"
```

## Testing

> **TL;DR**: Unit tests are fast, integration tests are slow but thorough. Use build tags to separate them.

### Test Structure

```
internal/
├── service/
│   └── service_test.go          # Unit tests (44 tests)
├── repository/
│   └── repository_test.go       # Unit tests  
├── models/
│   └── models_test.go           # Unit tests
└── tests/
    └── integration_test.go      # Integration tests (8 tests)
```

### Running Tests

```bash
# Fast unit tests only (default)
make test
# Runs in ~0.3s

# Integration tests only
make test-integration
# Runs in ~0.5s (SQLite) or ~0.8s (PostgreSQL)

# All tests (unit + integration)
make test-all

# Verbose output
make test VERBOSE=-v

# Specific test
go test ./internal/service -run TestCreateUser

# With coverage
make test-coverage
```

### Build Tags

Tests use Go build tags to separate unit and integration tests:

```go
//go:build integration

package tests
```

**Why?**
- Unit tests are fast (no real database)
- Integration tests are slow (real database, transactions)
- CI can run them separately

### Writing Unit Tests

Unit tests use mocked/in-memory databases:

```go
func TestService_CreateUser(t *testing.T) {
    db := testutil.SetupTestDB(t)  // In-memory SQLite
    repo := repository.New(db)
    svc := service.New(repo)
    
    user, err := svc.CreateUser(ctx, "test@example.com", ...)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    // Assertions...
}
```

### Writing Integration Tests

Integration tests use real databases with transaction rollback:

```go
//go:build integration

func TestUserCreation(t *testing.T) {
    database := testutil.SetupIntegrationDB(t)  // Real database
    
    testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
        queries := testutil.NewQueries(tx)
        repo := repository.NewWithQueries(queries)
        repo.SetDriverName(testutil.GetDriverName())
        svc := service.New(repo)
        
        // Test operations...
        
        // Transaction automatically rolls back after test
    })
}
```

**Transaction Rollback Pattern**:
- Each test runs in a transaction
- Transaction rolls back after test
- No database pollution
- Perfect isolation

### Testing Both Databases

```bash
# Test with SQLite (default)
unset TEST_DATABASE_URL
make test-integration

# Test with PostgreSQL
export TEST_DATABASE_URL="postgres://gosql:dev_password_123@localhost:5432/gosql_test?sslmode=disable"
make docker-up  # Start PostgreSQL first
make test-integration
```

## Database Development

> **TL;DR**: SQL files in `query/` and `migrations/`. Edit SQL → Run `sqlc generate` → Commit both.

### Directory Structure

```
query/                      # SQL queries for sqlc
├── users/
│   ├── users_read.sql     # SELECT queries
│   └── users_write.sql    # INSERT, UPDATE, DELETE
├── courses/
├── enrollments/
└── orders/

migrations/                 # Database migrations
├── 001_schema.sql         # SQLite initial schema
├── 002_add_orders.sql     # SQLite order system
└── postgres/              # PostgreSQL migrations
    ├── 001_schema.sql
    └── 002_add_orders.sql
```

### Adding a New Query

```sql
-- query/users/users_read.sql

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;
```

Then regenerate:
```bash
sqlc generate
```

This creates `internal/db/users_read.sql.go` with:
```go
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error)
```

### Adding a New Table

1. **Add migration SQL**:

```sql
-- migrations/003_add_notifications.sql (SQLite)
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
```

2. **Add PostgreSQL version**:

```sql
-- migrations/postgres/003_add_notifications.sql
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
```

3. **Add queries**:

```sql
-- query/notifications/notifications_read.sql

-- name: GetNotification :one
SELECT * FROM notifications WHERE id = ?;

-- name: ListUserNotifications :many
SELECT * FROM notifications
WHERE user_id = ?
ORDER BY created_at DESC;
```

4. **Generate code**:

```bash
sqlc generate
```

5. **Add to schema list** (update sqlc.yaml if needed):

```yaml
queries:
  - query/notifications  # Add this
```

### Database Migrations

Migrations run automatically when the application starts. They are idempotent (safe to run multiple times).

**SQLite migrations**:
- Located in `migrations/`
- Use `INTEGER PRIMARY KEY AUTOINCREMENT`
- Use `DATETIME`

**PostgreSQL migrations**:
- Located in `migrations/postgres/`
- Use `SERIAL PRIMARY KEY`
- Use `TIMESTAMP`

**Important**: Keep migrations in sync but adapted for each database.

### sqlc Configuration

```yaml
# sqlc.yaml
version: "2"
sql:
  - name: "sqlite"
    engine: "sqlite"
    schema: "migrations"
    queries:
      - query/users
      - query/courses
      - query/enrollments
      - query/orders
    gen:
      go:
        package: "db"
        out: "internal/db"
        emit_json_tags: true
        emit_interface: true
        sql_package: "database/sql"
```

### Query Syntax

**Supported annotations**:
- `:one` - Returns single row
- `:many` - Returns multiple rows
- `:exec` - No return value
- `:execrows` - Returns number of rows affected

**Parameters**:
```sql
-- Positional
SELECT * FROM users WHERE id = ?;

-- Named (with sqlc.arg)
SELECT * FROM users WHERE email = sqlc.arg('email');

-- Nullable (with sqlc.narg)
UPDATE users SET name = COALESCE(sqlc.narg('name'), name);
```

## Docker & PostgreSQL

> **TL;DR**: `make docker-up` starts PostgreSQL. `make docker-down` stops it. Use `.env` for config.

### Quick Start

```bash
# Initial setup
make setup              # Creates .env file

# Start PostgreSQL
make docker-up          # Starts postgres + pgAdmin

# Check status
make docker-ps

# Stop PostgreSQL
make docker-down

# Complete cleanup (removes volumes)
make docker-clean
```

### Environment Configuration

```bash
# .env file (created by make setup)
POSTGRES_USER=gosql
POSTGRES_PASSWORD=dev_password_123
POSTGRES_DB=gosql_dev
POSTGRES_HOST=localhost
POSTGRES_PORT=5432

# To use PostgreSQL in application
DATABASE_URL=postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable

# For tests
TEST_DATABASE_URL=postgres://gosql:dev_password_123@localhost:5432/gosql_test?sslmode=disable
```

### Docker Compose Services

**PostgreSQL**:
- Container: `gosql-postgres`
- Port: `5432`
- Volume: `gosql_postgres_data`
- Healthcheck: Automatic

**pgAdmin** (optional):
- Container: `gosql-pgadmin`
- Port: `5050`
- URL: `http://localhost:5050`
- Email: `admin@admin.com`
- Password: `admin`

### Useful Docker Commands

```bash
# View logs
make docker-logs

# PostgreSQL shell
make db-shell
# Inside psql:
\dt                    # List tables
\d users               # Describe table
SELECT * FROM users;
\q                     # Quit

# Restart services
make docker-restart

# Check health
docker compose ps
```

### Database Switching

```bash
# Use SQLite (default)
unset DATABASE_URL
./bin/gosql user-list
# Output: Connected to sqlite database

# Use PostgreSQL
export DATABASE_URL="postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable"
./bin/gosql user-list
# Output: Connected to postgres database
```

## Code Generation

> **TL;DR**: Run `sqlc generate` after editing SQL files. Don't edit generated files in `internal/db/`.

### What is sqlc?

sqlc generates type-safe Go code from SQL queries:

**Input** (SQL):
```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = ?;
```

**Output** (Go):
```go
func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
    row := q.db.QueryRowContext(ctx, getUser, id)
    var i User
    err := row.Scan(&i.ID, &i.Email, &i.Name, &i.Age, &i.CreatedAt)
    return i, err
}
```

### When to Regenerate

Run `sqlc generate` when you:
- ✅ Add/modify queries in `query/`
- ✅ Change database schema in `migrations/`
- ✅ Update `sqlc.yaml`
- ❌ DON'T manually edit `internal/db/`

### Verifying Generation

```bash
# Regenerate
sqlc generate

# Check git diff
git diff internal/db/

# If everything looks good
git add internal/db/ query/
git commit -m "feat: add new query"
```

## Metrics & Observability

> **TL;DR**: Built-in metrics wrapper ready for Prometheus, StatsD, or custom collectors. Zero overhead by default.

### Overview

The database layer includes a pluggable metrics wrapper that measures all database operations:

```
Application → MetricsWrapper → PostgresWrapper → sql.DB → Database
                     ↓
              MetricsCollector
```

**By default**: Uses `NoOpMetrics` (zero overhead, no-op)
**When needed**: Swap in Prometheus, StatsD, or custom implementation

### Architecture

The `MetricsWrapper` implements the same interface as `sql.DB`, so it can wrap any database connection:

```go
// internal/database/metrics.go

type MetricsCollector interface {
    ObserveQuery(duration time.Duration, err error, query string)
}

type MetricsWrapper struct {
    db      *sql.DB
    metrics MetricsCollector
}

func (m *MetricsWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    start := time.Now()
    res, err := m.db.ExecContext(ctx, query, args...)
    m.metrics.ObserveQuery(time.Since(start), err, query)
    return res, err
}
```

### Built-in Implementations

**NoOpMetrics** (default):
```go
type NoOpMetrics struct{}

func (m *NoOpMetrics) ObserveQuery(duration time.Duration, err error, query string) {
    // No-op: zero overhead
}
```

### Adding Custom Metrics

**Step 1**: Implement `MetricsCollector` interface

```go
// internal/database/prometheus_metrics.go (example)

import (
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    dbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_query_duration_seconds",
            Help: "Database query duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"query_type", "status"},
    )

    dbQueryTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_query_total",
            Help: "Total number of database queries",
        },
        []string{"query_type", "status"},
    )
)

type PrometheusMetrics struct{}

func (m *PrometheusMetrics) ObserveQuery(duration time.Duration, err error, query string) {
    queryType := extractQueryType(query)
    status := "success"
    if err != nil {
        status = "error"
    }

    dbQueryDuration.WithLabelValues(queryType, status).Observe(duration.Seconds())
    dbQueryTotal.WithLabelValues(queryType, status).Inc()
}

func extractQueryType(query string) string {
    query = strings.TrimSpace(strings.ToUpper(query))
    fields := strings.Fields(query)
    if len(fields) > 0 {
        return fields[0] // "SELECT", "INSERT", "UPDATE", "DELETE"
    }
    return "UNKNOWN"
}
```

**Step 2**: Update factory to use custom metrics

```go
// internal/database/factory.go

func newSQLiteConnection(ctx context.Context, dbPath string) (*Connection, error) {
    // ... existing setup code ...

    // Replace NoOpMetrics with your implementation
    metricsDB := NewMetricsWrapper(sqlDB, &PrometheusMetrics{})

    return &Connection{
        DB:      sqlDB,
        Queries: db.New(metricsDB),
        Driver:  "sqlite",
    }, nil
}
```

**Step 3**: Expose metrics endpoint (in main.go)

```go
import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Existing code...

    // Expose /metrics endpoint
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Fatal(http.ListenAndServe(":9090", nil))
    }()

    // Continue with CLI...
}
```

### Example: Logging Metrics (for debugging)

```go
// internal/database/log_metrics.go

import (
    "log"
    "strings"
    "time"
)

type LogMetrics struct{}

func (m *LogMetrics) ObserveQuery(duration time.Duration, err error, query string) {
    queryType := extractQueryType(query)
    status := "✓"
    if err != nil {
        status = "✗"
    }

    log.Printf("[DB] %s %s %v", status, queryType, duration)
}
```

### Metrics Data Points

The wrapper collects:
- **Duration**: How long each query took
- **Error**: Whether the query succeeded or failed
- **Query**: The SQL statement (for categorization)

Common metrics to derive:
- Query duration (histogram/percentiles)
- Query count (counter)
- Error rate (counter)
- Queries per second (rate)
- Slow queries (threshold-based alerts)

### Performance

**NoOpMetrics overhead**: ~0ns (compiler optimizes away)
**With metrics**: ~100-500ns per query (negligible for database operations)

### Testing Metrics

```go
// Example test
type TestMetrics struct {
    calls []MetricsCall
}

func (m *TestMetrics) ObserveQuery(duration time.Duration, err error, query string) {
    m.calls = append(m.calls, MetricsCall{Duration: duration, Err: err, Query: query})
}

func TestDatabaseMetrics(t *testing.T) {
    metrics := &TestMetrics{}
    db := NewMetricsWrapper(sqlDB, metrics)

    // Execute queries...

    if len(metrics.calls) != expectedCount {
        t.Errorf("expected %d queries, got %d", expectedCount, len(metrics.calls))
    }
}
```

See `internal/database/metrics_test.go` for complete examples.

## Common Tasks

> **TL;DR**: Quick reference for frequent development tasks.

### Adding a New CLI Command

```go
// cmd/app/main.go

type UserGetByEmailCmd struct {
    Email string `arg:"" help:"User email"`
}

func (cmd *UserGetByEmailCmd) Run(ctx context.Context, svc *service.Service) error {
    user, err := svc.GetUserByEmail(ctx, cmd.Email)
    if err != nil {
        return err
    }
    return printJSON(user)
}

// Add to CLI struct
type CLI struct {
    // ... existing commands
    UserGetByEmail UserGetByEmailCmd `cmd:"" help:"Get user by email"`
}
```

### Adding Validation

```go
// internal/service/service.go

func (s *Service) CreateUser(ctx context.Context, email string, name sql.NullString, age sql.NullInt64) (models.User, error) {
    // Validate email format
    if !strings.Contains(email, "@") {
        return models.User{}, fmt.Errorf("invalid email format")
    }

    // Validate age range
    if age.Valid && (age.Int64 < 0 || age.Int64 > 150) {
        return models.User{}, fmt.Errorf("age must be between 0 and 150")
    }

    user, err := s.repo.CreateUser(ctx, email, name, age)
    // ...
}
```

### Adding Transaction Logic

```go
// internal/repository/repository.go

func (r *Repository) ComplexOperation(ctx context.Context) error {
    return r.withTx(ctx, func(q *db.Queries) error {
        // Step 1
        if err := q.CreateSomething(ctx, ...); err != nil {
            return err  // Auto-rollback
        }

        // Step 2
        if err := q.UpdateSomething(ctx, ...); err != nil {
            return err  // Auto-rollback
        }

        // All steps succeeded, commit happens automatically
        return nil
    })
}
```

### Exporting Packages for Library Use

The repository and service packages can be used as a library:

```go
import (
    "example.com/go-sql/internal/database"
    "example.com/go-sql/internal/repository"
    "example.com/go-sql/internal/service"
)

func main() {
    ctx := context.Background()
    conn, _ := database.NewConnection(ctx)
    repo := repository.NewWithDB(conn.DB, conn.Queries, conn.Driver)
    svc := service.New(repo)

    user, _ := svc.CreateUser(ctx, "user@example.com", ...)
}
```

## Troubleshooting

> **TL;DR**: Common issues and solutions.

### sqlc Issues

**Problem**: `sqlc: command not found`
```bash
# Install sqlc (standard Go way)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

# Verify installation
sqlc version
```

**Problem**: sqlc generation fails
```bash
# Check syntax in SQL files
cat query/users/users_read.sql

# Validate sqlc.yaml
sqlc verify

# Try with verbose output
sqlc generate -f sqlc.yaml
```

### PostgreSQL Issues

**Problem**: Can't connect to PostgreSQL
```bash
# Check if running
make docker-ps

# Start if not running
make docker-up

# Check logs
make docker-logs

# Test connection
psql "postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable"
```

**Problem**: Migration fails
```bash
# Check migration files exist
ls migrations/postgres/

# Check syntax
cat migrations/postgres/001_schema.sql

# Test manually
make db-shell
\i /path/to/migration.sql
```

**Problem**: Wrong database selected
```bash
# Check environment
echo $DATABASE_URL

# Should be:
# postgres://...  for PostgreSQL
# or empty/file path for SQLite
```

### Test Issues

**Problem**: Integration tests fail
```bash
# Check database
make docker-ps  # For PostgreSQL

# Check environment
echo $TEST_DATABASE_URL

# Run with verbose output
go test -tags=integration ./internal/tests/... -v

# Clean and retry
make docker-down
make docker-up
make test-integration
```

**Problem**: Tests pass individually but fail together
```bash
# Likely a transaction isolation issue
# Check that tests use WithTxContext
# Each test should be independent
```

### Build Issues

**Problem**: Build fails after SQL changes
```bash
# Regenerate code
sqlc generate

# Clean and rebuild
make clean
make build
```

**Problem**: Import cycle
```bash
# Check dependencies
# Layers should be: CLI → Service → Repository → Database
# If Repository imports Service, you have a cycle
```

## Performance Tips

1. **Use prepared statements** for bulk operations
2. **Index foreign keys** (already done in migrations)
3. **Use transactions** for multi-step operations
4. **Connection pooling** (PostgreSQL: 25, SQLite: 1)
5. **Pagination** for large result sets
6. **LEFT JOIN** instead of multiple queries

## Code Style

- Follow Go conventions (`gofmt`, `golint`)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused
- Write tests for new features
- Update documentation

## Contributing Workflow

```bash
# 1. Create issue (describe feature/bug)
# 2. Create branch
git checkout -b feature/issue-123

# 3. Develop with tests
make test

# 4. Commit with conventional commits
git commit -m "feat: add user search"
git commit -m "fix: handle null emails"
git commit -m "docs: update README"

# 5. Push and create PR
git push origin feature/issue-123

# 6. Address review feedback
# 7. Merge when approved
```

## Resources

- **[sqlc Documentation](https://docs.sqlc.dev/)** - Query generation
- **[Kong Documentation](https://github.com/alecthomas/kong)** - CLI framework
- **[Go Testing](https://golang.org/pkg/testing/)** - Testing guide
- **[PostgreSQL Docs](https://www.postgresql.org/docs/)** - PostgreSQL reference
- **[SQLite Docs](https://www.sqlite.org/docs.html)** - SQLite reference

