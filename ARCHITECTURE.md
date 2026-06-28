# Architecture

> **TL;DR**: Clean Architecture with 3 layers: CLI → Service → Repository → Database. Multi-database support (SQLite/PostgreSQL) via factory pattern. Type-safe SQL with sqlc. Transaction-based operations for data integrity.

## Table of Contents

- [High-Level Overview](#high-level-overview)
- [Project Structure](#project-structure)
- [Architecture Layers](#architecture-layers)
- [Database Support](#database-support)
- [Data Flow](#data-flow)
- [Key Design Patterns](#key-design-patterns)
- [Database Schema](#database-schema)

## High-Level Overview

This is a CLI tool for managing an e-learning platform, built following Clean Architecture principles.

The architecture consists of 4 distinct layers, each with a single responsibility:

```
┌─────────────────────────────────────────────┐
│           CLI Layer (Kong)                  │
│  - Command parsing                          │
│  - JSON output                              │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│         Service Layer                       │
│  - Business logic                           │
│  - Error translation                        │
│  - Transaction orchestration                │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│        Repository Layer                     │
│  - Data access                              │
│  - Transaction management                   │
│  - Database abstraction                     │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│    Database Layer (sqlc generated)          │
│  - Type-safe SQL queries                    │
└──────────────┬──────────────────────────────┘
               │
               ▼
       ┌───────┴───────┐
       │               │
  ┌────▼────┐     ┌────▼────────┐
  │ SQLite  │     │ PostgreSQL  │
  └─────────┘     └─────────────┘
```

**Dependencies point inward**: CLI depends on Service, Service depends on Repository, Repository depends on Database. Database layer has no dependencies.

## Project Structure

> **TL;DR**: `cmd/` = CLI, `internal/` = business logic, `query/` = SQL files, `migrations/` = schema.

```
.
├── cmd/app/
│   └── main.go                   # CLI entry point & Kong commands
│
├── internal/
│   ├── database/
│   │   └── factory.go            # Multi-database factory
│   │
│   ├── db/                       # sqlc generated (DON'T EDIT!)
│   │   ├── querier.go
│   │   ├── models.go
│   │   └── *.sql.go
│   │
│   ├── models/
│   │   └── models.go             # Domain models
│   │
│   ├── repository/
│   │   ├── repository.go         # Data access layer
│   │   └── errors.go             # Domain errors
│   │
│   ├── service/
│   │   └── service.go            # Business logic
│   │
│   └── testutil/
│       ├── testutil.go           # Test helpers
│       └── integration.go
│
├── query/                        # SQL files (for sqlc)
│   ├── users/
│   ├── courses/
│   ├── enrollments/
│   └── orders/
│
├── migrations/
│   ├── 001_schema.sql            # SQLite
│   ├── 002_add_orders.sql
│   └── postgres/                 # PostgreSQL
│       ├── 001_schema.sql
│       └── 002_add_orders.sql
│
├── sqlc.yaml                     # sqlc config
├── Makefile
└── docker-compose.yml
```

## Architecture Layers

> **TL;DR**: Each layer has one job. CLI handles I/O, Service has business rules, Repository accesses data, Database executes SQL.

### 1. CLI Layer (`cmd/app/main.go`)

**Responsibility**: User interaction

```go
func (cmd *UserAddCmd) Run(ctx context.Context, svc *service.Service) error {
    user, err := svc.CreateUser(ctx, cmd.Email, name, age)
    if err != nil {
        return err  // Service provides user-friendly messages
    }
    return printJSON(user)
}
```

**Rules**:
- ✅ Parse commands with Kong
- ✅ Format output as JSON
- ✅ Call service methods
- ❌ NO business logic
- ❌ NO database access

### 2. Service Layer (`internal/service/`)

**Responsibility**: Business logic and orchestration

```go
func (s *Service) CreateUser(ctx context.Context, email string, name sql.NullString, age sql.NullInt64) (models.User, error) {
    user, err := s.repo.CreateUser(ctx, email, name, age)
    if err != nil {
        if repository.IsConflict(err) {
            return models.User{}, fmt.Errorf("user with email '%s' already exists", email)
        }
        return models.User{}, fmt.Errorf("failed to create user: %w", err)
    }
    return models.FromDBUser(user), nil
}
```

**Rules**:
- ✅ Validate business rules
- ✅ Orchestrate repository calls
- ✅ Translate errors to user messages
- ✅ Return domain models
- ❌ NO direct SQL

### 3. Repository Layer (`internal/repository/`)

**Responsibility**: Data access abstraction

```go
func (r *Repository) CreateUser(ctx context.Context, email string, name sql.NullString, age sql.NullInt64) (db.User, error) {
    user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
        Email: email,
        Name:  name,
        Age:   age,
    })
    return user, mapError(err)  // Translate DB errors to domain errors
}
```

**Domain Errors**:
- `ErrNotFound` - Record doesn't exist
- `ErrConflict` - Constraint violation (unique, foreign key)

**Transaction Support**:
```go
func (r *Repository) CreateOrderWithEnrollments(ctx context.Context, userID int64, courseIDs []int64, paymentMethod string) (db.Order, error) {
    var order db.Order
    err := r.withTx(ctx, func(q *db.Queries) error {
        // Multiple operations in one transaction
        // Automatic rollback on error
    })
    return order, err
}
```

### 4. Database Layer (`internal/db/` - sqlc generated)

**Responsibility**: Type-safe SQL execution

Generated from SQL files with `sqlc generate`. **DO NOT EDIT** these files manually.

**Example generated code**:
```go
const createUser = `INSERT INTO users (email, name, age) VALUES (?, ?, ?) RETURNING *`

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
    row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.Name, arg.Age)
    var i User
    err := row.Scan(&i.ID, &i.Email, &i.Name, &i.Age, &i.CreatedAt)
    return i, err
}
```

## Database Support

> **TL;DR**: Factory pattern detects `DATABASE_URL` and creates appropriate connection. Placeholder converter makes SQLite queries work on PostgreSQL.

### Multi-Database Strategy

One codebase supports two databases:

```
Environment Variable (DATABASE_URL)
            ↓
     Factory Detection
            ↓
    ┌───────┴───────┐
    │               │
    ▼               ▼
SQLite          PostgreSQL
    ↓               ↓
No wrapper      Placeholder wrapper
                (? → $1, $2, $3)
    ↓               ↓
    └───────┬───────┘
            ↓
    Repository Layer
            ↓
     Service Layer
            ↓
       CLI Layer
```

### Database Factory

```go
func NewConnection(ctx context.Context) (*Connection, error) {
    dbURL := os.Getenv("DATABASE_URL")

    if strings.HasPrefix(dbURL, "postgres://") {
        return newPostgresConnection(ctx, dbURL)
    }

    // Default to SQLite
    return newSQLiteConnection(ctx, dbPath)
}
```

### Placeholder Conversion

**Challenge**: SQLite uses `?`, PostgreSQL uses `$1, $2, $3`

**Solution**: Runtime wrapper converts placeholders for PostgreSQL

```go
type postgresDBWrapper struct {
    db *sql.DB
}

func (w *postgresDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    // "SELECT * FROM users WHERE id = ?" → "SELECT * FROM users WHERE id = $1"
    converted := convertPlaceholders(query)
    return w.db.QueryContext(ctx, converted, args...)
}
```

This allows us to:
- Write SQL queries once (using `?`)
- Generate code once with sqlc
- Support both databases at runtime

### Database-Specific Features

| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| **Placeholders** | `?` | `$1, $2, $3` |
| **Auto-increment** | `AUTOINCREMENT` | `SERIAL` |
| **JSON aggregation** | `json_group_array()` | `json_agg()` |
| **Timestamp** | `DATETIME` | `TIMESTAMP` |
| **Connection pool** | 1 (single writer) | 25 (concurrent) |
| **Constraint errors** | Message parsing | Error codes (23505, 23503) |

## Data Flow

> **TL;DR**: Request flows down through layers, response flows back up. Transactions ensure atomic operations.

### Example: Creating an Order with Enrollments

```
User Input
    ↓
1. CLI parses command
   ./bin/gosql order-create -u 1 -c 1 -c 2 -p card
    ↓
2. CLI calls Service
   svc.CreateOrder(ctx, 1, []int64{1,2}, "card")
    ↓
3. Service calls Repository
   repo.CreateOrderWithEnrollments(ctx, 1, []int64{1,2}, "card")
    ↓
4. Repository starts transaction
   BEGIN TRANSACTION
    ↓
5. Validate user exists
   SELECT 1 FROM users WHERE id = 1
    ↓
6. Get course prices
   SELECT price FROM courses WHERE id IN (1, 2)
    ↓
7. Create order
   INSERT INTO orders ...
    ↓
8. Create order items
   INSERT INTO order_items ...
    ↓
9. Mark order complete
   UPDATE orders SET status = 'completed' ...
    ↓
10. Create enrollments
    INSERT INTO enrollments ...
    ↓
11. Commit transaction
    COMMIT
    ↓
12. Return domain model
    Repository → Service → CLI
    ↓
13. Output JSON
    {"id": 1, "status": "completed", ...}
```

**If any step fails**: Transaction automatically rolls back, no partial data.

## Key Design Patterns

> **TL;DR**: Repository, Factory, Transaction Script, Domain Errors, Dependency Injection.

### 1. Repository Pattern

Encapsulates all data access for each entity:

```go
type Repository struct {
    db      *sql.DB
    queries *db.Queries
}

// All user operations
func (r *Repository) CreateUser(...)
func (r *Repository) GetUser(...)
func (r *Repository) ListUsers(...)

// All course operations
func (r *Repository) CreateCourse(...)
// ...
```

**Benefits**:
- Single source of truth for queries
- Easy to mock for testing
- Database-agnostic interface

### 2. Factory Pattern

Creates database connections based on configuration:

```go
type Connection struct {
    DB      *sql.DB
    Queries *db.Queries
    Driver  string  // "sqlite" or "postgres"
}

func NewConnection(ctx context.Context) (*Connection, error)
```

### 3. Transaction Script

Complex operations wrapped in transactions:

```go
func (r *Repository) withTx(ctx context.Context, fn func(*db.Queries) error) error {
    tx, _ := r.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    if err := fn(db.New(tx)); err != nil {
        return err  // Rollback happens in defer
    }

    return tx.Commit()
}
```

### 4. Domain Errors

Abstract database errors to domain concepts:

```go
var (
    ErrNotFound = errors.New("not found")
    ErrConflict = errors.New("conflict")
)

func mapError(err error) error {
    if errors.Is(err, sql.ErrNoRows) {
        return ErrNotFound
    }
    if strings.Contains(err.Error(), "UNIQUE constraint") ||
       strings.Contains(err.Error(), "23505") {
        return ErrConflict
    }
    return err
}
```

## Database Schema

> **TL;DR**: 5 tables (users, courses, enrollments, orders, order_items) with foreign keys and unique constraints.

### Entity-Relationship Diagram

```
┌────────────┐
│   users    │
└──────┬─────┘
       │
       ├─────────────────┐
       │                 │
       ▼                 ▼
┌────────────┐    ┌──────────────┐
│   orders   │    │ enrollments  │
└──────┬─────┘    └──────┬───────┘
       │                 │
       ▼                 │
┌───────────────┐        │
│  order_items  │        │
└──────┬────────┘        │
       │                 │
       │    ┌────────────┘
       ▼    ▼
    ┌──────────┐
    │ courses  │
    └──────────┘
```

### Tables

**users**
- `id` - Primary key
- `email` - UNIQUE, required
- `name` - Nullable
- `age` - Nullable
- `created_at` - Timestamp

**courses**
- `id` - Primary key
- `slug` - UNIQUE, required
- `title` - Required
- `price` - Integer (cents), default 0

**enrollments**
- `id` - Primary key
- `user_id` - Foreign key to users
- `course_id` - Foreign key to courses
- `enrolled_at` - Timestamp
- `status` - active/completed/cancelled
- `order_id` - Foreign key to orders (nullable, for free enrollments)
- UNIQUE(user_id, course_id)

**orders**
- `id` - Primary key
- `user_id` - Foreign key to users
- `total_amount` - Integer (cents)
- `status` - pending/completed/failed/refunded
- `payment_method` - Nullable
- `created_at` - Timestamp
- `completed_at` - Nullable timestamp

**order_items**
- `id` - Primary key
- `order_id` - Foreign key to orders
- `course_id` - Foreign key to courses
- `price` - Integer (historical price at purchase)
- `created_at` - Timestamp
- UNIQUE(order_id, course_id)

### Indexes

All foreign keys are indexed:
- `idx_orders_user_id`
- `idx_orders_status`
- `idx_order_items_order_id`
- `idx_order_items_course_id`
- `idx_enrollments_order_id`

## Technology Stack

- **Language**: Go 1.25+
- **CLI Framework**: Kong (command parsing)
- **SQL Generator**: sqlc v1.31.1 (type-safe queries)
- **Databases**: SQLite (default), PostgreSQL (optional)
- **Testing**: Go testing package with build tags
- **CI/CD**: GitHub Actions (future)
- **Containers**: Docker & Docker Compose

## Design Principles

1. **Separation of Concerns**: Each layer has one responsibility
2. **Dependency Inversion**: Dependencies point inward
3. **Interface Segregation**: Small, focused interfaces (Querier)
4. **Single Source of Truth**: SQL files define schema and queries
5. **Fail Fast**: Validate early, fail with clear errors
6. **Transaction Safety**: Atomic operations for data integrity
7. **Type Safety**: Compile-time checks via sqlc
8. **Testability**: Easy to mock each layer

## See Also

- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Developer guide
- **[README.md](README.md)** - User documentation

