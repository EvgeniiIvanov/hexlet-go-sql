# Go SQL CLI Tool

> **TL;DR**: A production-ready CLI for managing an e-learning platform with courses, users, and orders. Supports both SQLite and PostgreSQL with automatic database switching. Built with type-safe SQL (sqlc) and Clean Architecture.

A command-line interface for managing an **e-learning platform** with **SQLite** or **PostgreSQL**. Supports courses, users, enrollments, and a complete **order/payment system**. Built with **sqlc** for type-safe SQL queries and following **Clean Architecture** principles.

## Features

### Core Features
- **E-Commerce System**: Complete order management with purchase tracking and enrollments
- **Multi-Database**: SQLite (default) or PostgreSQL - switch with one environment variable
- **Type-Safe SQL**: sqlc generates Go code from SQL queries
- **Clean Architecture**: Service → Repository → Database layer separation
- **Transactions**: Atomic multi-step operations with proper error handling
- **JSON Output**: All commands output structured JSON
- **Comprehensive Tests**: 52 tests (44 unit + 8 integration) with 23-54% coverage

### Database Support
- **SQLite**: Zero-config, perfect for development and single-user deployments
- **PostgreSQL**: Production-ready with Docker support and connection pooling
- **Automatic Switching**: Set `DATABASE_URL` environment variable to switch databases
- **Migrations**: Automatic schema migration on startup for both databases

## Requirements

- **Go 1.25.0 or higher**
- **sqlc v1.31.1** - SQL to Go code generator
- **goose v3.27.1** - Database migration tool
- Docker & Docker Compose (optional, for PostgreSQL)

```bash
# Install sqlc (standard Go way)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

# Install goose (standard Go way)
go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

# Verify installations
sqlc version
goose -version
```

## Quick Start

> **TL;DR**: `make build && ./bin/gosql user-add -e "alice@example.com" -n "Alice"` - that's it!

### Option 1: SQLite (Default - No Setup Required!)

```bash
# Clone and build
git clone <repo-url>
cd hexlet-go-sql
make build

# Start using immediately (no database setup needed!)
./bin/gosql user-add -e "alice@example.com" -n "Alice"
```

### Option 2: PostgreSQL with Docker

```bash
# Clone and setup
git clone <repo-url>
cd hexlet-go-sql
make setup         # Creates .env file
make docker-up     # Starts PostgreSQL in Docker

# Build and use
make build
export DATABASE_URL="postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable"
./bin/gosql user-add -e "alice@example.com" -n "Alice"
```

### Switching Between Databases

```bash
# Use SQLite (default)
unset DATABASE_URL
./bin/gosql user-list

# Use PostgreSQL
export DATABASE_URL="postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable"
./bin/gosql user-list
```

## Usage

### Course Commands

**Add a Course:**
```bash
./bin/gosql course-add -s "go-basics" -t "Go Programming Basics" -p 100
```

**List Courses:**
```bash
./bin/gosql course-list
./bin/gosql course-list -l 5 -o 0   # First 5 courses
./bin/gosql course-list -l 5 -o 5   # Next 5 courses (pagination)
```

**Find Courses by IDs:**
```bash
./bin/gosql course-find 1 2 3
```

**Bulk Upsert Courses from JSON file:**
```bash
./bin/gosql course-bulk-upsert -f examples/courses.json
```

**Bulk Upsert Courses from stdin:**
```bash
echo '[{"slug":"go","title":"Go Course","price":100}]' | ./bin/gosql course-bulk-upsert
```

### User Commands

**Add a User (with all fields):**
```bash
./bin/gosql user-add -e "john@example.com" -n "John Doe" -a 30
```

**Add a User (nullable fields as NULL):**
```bash
./bin/gosql user-add -e "anonymous@example.com"
```
Note: `name` and `age` are omitted from JSON output when NULL.

**List All Users:**
```bash
./bin/gosql user-list
./bin/gosql user-list -l 20 -o 0   # First 20 users
./bin/gosql user-list -l 20 -o 20  # Next 20 users (pagination)
```

**Get User by ID:**
```bash
./bin/gosql user-get 1
```

**Bulk Upsert Users from JSON file:**
```bash
./bin/gosql user-bulk-upsert -f examples/users.json
```

**Bulk Upsert Users from stdin:**
```bash
echo '[{"email":"user@example.com","name":"John","age":30}]' | ./bin/gosql user-bulk-upsert
```

### Enrollment Commands

Enrollment operations use **transactions** to ensure data integrity. All validation happens within a transaction, which rolls back automatically if any check fails.

**Enroll a User in a Course:**
```bash
./bin/gosql enrollment-create -u 1 -c 1
```
This uses a transaction to:
1. Verify the user exists
2. Verify the course exists
3. Create the enrollment
4. Rollback if ANY step fails

**List All Enrollments:**
```bash
./bin/gosql enrollment-list
./bin/gosql enrollment-list -l 10 -o 0  # First 10 enrollments
```

**List Enrollments by User:**
```bash
./bin/gosql enrollment-by-user -u 1
```

**List Enrollments by Course:**
```bash
./bin/gosql enrollment-by-course -c 1
```

**Complete an Enrollment:**
```bash
./bin/gosql enrollment-complete -u 1 -c 1
```

**Cancel an Enrollment:**
```bash
./bin/gosql enrollment-cancel -u 1 -c 1
```

### Order Commands (E-Commerce)

**Create an Order (Purchase Courses):**
```bash
# Purchase a single course
./bin/gosql order-create -u 1 -c 1 -p card

# Purchase multiple courses in one order
./bin/gosql order-create -u 1 -c 1 -c 2 -c 3 -p paypal
```

This uses a **transaction** to atomically:
1. Verify user and all courses exist
2. Calculate total amount from course prices
3. Create ORDER record (pending → completed)
4. Create ORDER_ITEMS for each course (with historical prices)
5. Create ENROLLMENTS for each course (linked to order)
6. Rollback if ANY step fails

Output:
```json
{
  "id": 1,
  "user_id": 1,
  "total_amount": 24998,
  "status": "completed",
  "payment_method": "card",
  "created_at": "2026-06-26T10:19:55Z"
}
```

**Get Order Details:**
```bash
./bin/gosql order-get 1
```

Shows order with all purchased courses (using `json_group_array`):
```json
{
  "id": 1,
  "total_amount": 24998,
  "items": [
    {"course_title": "Go Programming", "price": 9999},
    {"course_title": "Rust Programming", "price": 14999}
  ]
}
```

**List All Orders:**
```bash
./bin/gosql order-list
./bin/gosql order-list -l 10 -o 0  # Pagination
```

**Get User's Order History:**
```bash
./bin/gosql orders-by-user -u 1
```

**Check Course Access/Ownership:**
```bash
./bin/gosql check-course-access -u 1 -c 1
```

Output:
```json
{
  "user_id": 1,
  "course_id": 1,
  "has_access": true
}
```

**Refund an Order:**
```bash
./bin/gosql order-refund 1
```

### Advanced Queries

**Get Course with All Enrollments (JSON aggregation):**
```bash
./bin/gosql course-with-enrollments 1
```

Output shows course with nested enrollment data:
```json
{
  "id": 1,
  "slug": "go-101",
  "title": "Go Programming",
  "price": 9999,
  "enrollments": [
    {
      "user_email": "alice@example.com",
      "enrolled_at": "2026-06-26 10:08:37",
      "status": "active"
    }
  ]
}
```

## Demo Scripts

### Transaction Demo

Run the enrollment demo to see transactions in action:
```bash
./examples/enrollment_demo.sh
```

This demonstrates:
- ✅ Successful enrollments (transaction commits)
- ❌ Failed enrollments with non-existent users (transaction rolls back)
- ❌ Failed enrollments with non-existent courses (transaction rolls back)
- ❌ Duplicate enrollment prevention (UNIQUE constraint + rollback)
- 📊 Enrollment status tracking (active, completed, cancelled)

Key transaction benefits:
- **Atomicity**: All checks pass or nothing happens
- **Consistency**: UNIQUE constraints prevent duplicates
- **Integrity**: Foreign keys prevent orphaned records
- **Rollback**: Invalid operations have no side effects

## Performance Demo

Run the bash script to see prepared statements in action:
```bash
./examples/bulk_performance_demo.sh
```

This demonstrates:
- Bulk insert of 100 users
- Bulk update of 50 existing users
- Bulk insert of 20 courses
- Mixed upsert operations

Each operation shows timing metrics including:
- `operation_time` - Time spent in database operations
- `total_time` - Total command execution time
- `avg_per_record` - Average time per record

Example output:
```json
{
  "avg_per_record": "649.82µs",
  "count": 100,
  "message": "Successfully upserted 100 users",
  "operation_time": "64.982ms",
  "success": true,
  "total_time": "65.097292ms"
}
```

## Example JSON Output

**User:**
```json
{
  "id": 1,
  "email": "john@example.com",
  "name": "John Doe",
  "age": 30,
  "created_at": "2026-06-21T09:58:05Z"
}
```

**Enrollment with Details:**
```json
{
  "id": 1,
  "user_id": 1,
  "user_email": "alice@example.com",
  "user_name": "Alice Smith",
  "course_id": 1,
  "course_slug": "go-basics",
  "course_title": "Go Programming Basics",
  "enrolled_at": "2026-06-24T10:52:13Z",
  "status": "active"
}
```

## Project Structure

```
.
├── cmd/app/
│   └── main.go                  # CLI commands and entry point
│
├── internal/
│   ├── database/
│   │   └── database.go          # Database connection and migrations
│   │
│   ├── db/                      # sqlc-generated code (DO NOT EDIT MANUALLY)
│   │   ├── db.go                # Database interface and Queries struct
│   │   ├── models.go            # Auto-generated models (Order, OrderItem, etc.)
│   │   ├── querier.go           # Interface for all queries
│   │   ├── courses_read.sql.go  # Generated course read queries
│   │   ├── courses_write.sql.go # Generated course write queries
│   │   ├── users_read.sql.go    # Generated user read queries
│   │   ├── users_write.sql.go   # Generated user write queries
│   │   ├── enrollments_read.sql.go  # Generated enrollment read queries
│   │   ├── enrollments_write.sql.go # Generated enrollment write queries
│   │   ├── orders_read.sql.go   # Generated order read queries
│   │   └── orders_write.sql.go  # Generated order write queries
│   │
│   ├── models/
│   │   └── models.go            # JSON-friendly model converters (sql.Null* → pointers)
│   │
│   ├── repository/
│   │   ├── errors.go            # Domain errors (ErrNotFound, ErrConflict)
│   │   └── repository.go        # Data access layer with error mapping
│   │
│   └── service/
│       └── service.go           # Business logic layer
│
├── migrations/
│   ├── 001_schema.sql           # Core schema (users, courses, enrollments)
│   └── 002_add_orders.sql       # Order system (orders, order_items, indexes)
│
├── query/
│   ├── courses/
│   │   ├── courses_read.sql     # Read queries (SELECT, json_group_array)
│   │   └── courses_write.sql    # Write queries (INSERT, UPDATE, DELETE)
│   ├── users/
│   │   ├── users_read.sql       # Read queries (SELECT, pagination)
│   │   └── users_write.sql      # Write queries (INSERT, UPDATE, DELETE)
│   ├── enrollments/
│   │   ├── enrollments_read.sql # Read queries (SELECT with LEFT JOIN)
│   │   └── enrollments_write.sql # Write queries (INSERT, UPDATE)
│   └── orders/
│       ├── orders_read.sql      # Read queries (SELECT with json_group_array)
│       └── orders_write.sql     # Write queries (INSERT, UPDATE)
│
├── examples/
│   ├── enrollment_demo.sh       # Transaction demo script
│   ├── bulk_performance_demo.sh # Bulk operations demo
│   ├── README_TRANSACTIONS.md   # Transaction implementation guide
│   └── *.json                   # Sample data files
│
├── sqlc.yaml                    # sqlc configuration
└── Makefile                     # Build commands
```

## Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                               │
│                      (cmd/app/main.go)                          │
│  • Parses commands                                              │
│  • Formats JSON output                                          │
│  • NO business logic or database access                         │
└────────────────────────────┬────────────────────────────────────┘
                             │ calls
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                              │
│                (internal/service/service.go)                    │
│  • Business logic                                               │
│  • Error handling with clear messages                           │
│  • Model conversion (db → JSON-friendly)                        │
│  • Orchestrates repository operations                           │
└────────────────────────────┬────────────────────────────────────┘
                             │ calls
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Repository Layer                             │
│              (internal/repository/repository.go)                │
│  • Wraps sqlc queries                                           │
│  • Transaction management                                       │
│  • Error mapping (sql → domain errors)                          │
│  • Bulk operations                                              │
└────────────────────────────┬────────────────────────────────────┘
                             │ calls
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Database Layer (sqlc)                        │
│                      (internal/db/*.go)                         │
│  • Type-safe SQL operations                                     │
│  • Generated from SQL queries                                   │
│  • Zero runtime overhead                                        │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities

**1. CLI Layer (`cmd/app/main.go`)**
- Handles user input/output
- Parses command-line flags
- Formats JSON responses
- Does NOT contain business logic or database calls

**2. Service Layer (`internal/service/service.go`)**
- Contains business logic
- Handles domain-specific errors with clear messages
- Converts database models to JSON-friendly types
- Orchestrates repository operations

**3. Repository Layer (`internal/repository/repository.go`)**
- Wraps sqlc-generated queries
- Manages transactions
- Maps database errors to domain errors (`ErrNotFound`, `ErrConflict`)
- Provides bulk operations

**4. Database Layer (`internal/db/` - generated by sqlc)**
- Type-safe SQL operations
- Generated from SQL queries
- Zero runtime overhead

### Error Handling

Domain-specific errors are defined in `internal/repository/errors.go`:

```go
// Repository errors
var (
    ErrNotFound = errors.New("resource not found")
    ErrConflict = errors.New("resource already exists")
)
```

**Error flow:**
```
Database Error (sql.ErrNoRows)
    ↓
Repository (maps to ErrNotFound)
    ↓
Service (adds context: "user with id 5 not found")
    ↓
CLI (displays error to user)
```

**Example:**
```go
// Service layer
func (s *Service) GetUser(ctx context.Context, id int64) (models.User, error) {
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        if repository.IsNotFound(err) {
            return models.User{}, fmt.Errorf("user with id %d not found", id)
        }
        return models.User{}, fmt.Errorf("failed to get user: %w", err)
    }
    return models.FromDBUser(user), nil
}
```

### sqlc: Type-Safe SQL

This project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries.

**Benefits:**
- ✅ Compile-time SQL validation
- ✅ Type-safe database operations
- ✅ No ORM overhead
- ✅ Full SQL feature support
- ✅ Automatic code generation

**How it works:**

1. **Write SQL schemas** in `migrations/*.sql`
2. **Write SQL queries** in `query/<entity>/{read|write}.sql`
3. **Generate type-safe Go code**: `sqlc generate`
4. **Use via Repository/Service layers** (not directly in CLI)

### Model Converters

Since sqlc generates models with `sql.Null*` types (for nullable fields), we use converter functions in `internal/models/models.go` to transform them into clean JSON-friendly types:

```go
// sqlc generates this:
type User struct {
    Name sql.NullString  // {"String": "John", "Valid": true}
}

// We convert to this:
type User struct {
    Name *string  // "John" or omitted if null
}
```

**Conversion happens in Service layer:**
```go
// Service layer
func (s *Service) GetUser(ctx context.Context, id int64) (models.User, error) {
    dbUser, err := s.repo.GetUser(ctx, id)  // Returns db.User (with sql.Null*)
    if err != nil {
        return models.User{}, err
    }
    return models.FromDBUser(dbUser), nil  // Convert to models.User (with pointers)
}
```

**CLI just calls service:**
```go
// CLI layer
func (cmd *UserGetCmd) Run(ctx context.Context, svc *service.Service) error {
    user, err := svc.GetUser(ctx, cmd.ID)  // Already converted to models.User
    if err != nil {
        return err
    }
    return printJSON(user)  // Clean JSON output
}
```

### Data Flow Example

Here's how a complete operation flows through the layers:

```
1. CLI receives command:
   ./bin/gosql user-get 1

2. CLI calls Service:
   user, err := svc.GetUser(ctx, 1)

3. Service calls Repository:
   dbUser, err := repo.GetUser(ctx, 1)

4. Repository calls sqlc Query (with error mapping):
   user, err := queries.GetUser(ctx, 1)
   return user, mapError(err)  // sql.ErrNoRows → ErrNotFound

5. Service converts model:
   return models.FromDBUser(dbUser), nil

6. CLI outputs JSON:
   return printJSON(user)
```

### Transactions

Transactions ensure data integrity for multi-step operations:

```go
func (r *Repository) EnrollUser(ctx context.Context, userID, courseID int64) (db.Enrollment, error) {
    var enrollment db.Enrollment

    err := r.withTx(ctx, func(q *db.Queries) error {
        // 1. Check user exists
        _, err := q.CheckUserExists(ctx, userID)
        if err != nil {
            return fmt.Errorf("user not found")
        }

        // 2. Check course exists
        _, err = q.CheckCourseExists(ctx, courseID)
        if err != nil {
            return fmt.Errorf("course not found")
        }

        // 3. Create enrollment
        enrollment, err = q.CreateEnrollment(ctx, ...)
        return err
    })

    return enrollment, err  // All or nothing
}
```

**Transaction guarantees:**
- All operations succeed or all fail (atomicity)
- Database constraints enforced (consistency)
- No partial state (isolation)
- Changes are permanent once committed (durability)

### Order System Architecture

The order system implements a complete e-commerce flow with transactional guarantees:

**Database Schema:**
```
orders
  ├── id, user_id, total_amount, status, payment_method
  └── timestamps (created_at, completed_at)

order_items
  ├── id, order_id, course_id, price
  └── Historical price tracking (price at time of purchase)

enrollments
  ├── ... (existing fields)
  └── order_id  -- Links enrollment to purchase (NULL for free courses)
```

**Purchase Flow (Single Transaction):**
```go
func CreateOrderWithEnrollments(ctx, userID, courseIDs, paymentMethod) {
    BEGIN TRANSACTION
        1. Verify user exists (fail fast)
        2. Get all course prices & verify courses exist
        3. Calculate total_amount
        4. CREATE order (status='pending')
        5. CREATE order_items (one per course, with historical price)
        6. UPDATE order (status='completed')
        7. CREATE enrollments (one per course, linked to order_id)
    COMMIT  -- All or nothing!
}
```

**Key Benefits:**
- **Atomic purchases**: User gets all courses or none (no partial orders)
- **Price history**: Track what price was paid (even if course price changes)
- **Audit trail**: Every enrollment links back to its order
- **Idempotent**: UNIQUE constraint prevents duplicate course purchases
- **Defensive**: LEFT JOIN handles deleted courses gracefully

**JSON Aggregation:**
Uses SQLite's `json_group_array()` to efficiently return nested data:
```sql
SELECT
    o.id, o.total_amount,
    json_group_array(json_object(
        'course_title', c.title,
        'price', oi.price
    )) as items
FROM orders o
LEFT JOIN order_items oi ON oi.order_id = o.id
LEFT JOIN courses c ON c.id = oi.course_id
GROUP BY o.id;
```

Result: Single query returns order with all purchased courses as JSON array.

## Quick Reference

### Price Handling

All prices are stored as **integers in cents** to avoid floating-point precision issues:

```bash
# $99.99 is stored as 9999 cents
./bin/gosql course-add -s "go-101" -t "Go Course" -p 9999

# Output shows:
{
  "price": 9999  # = $99.99
}
```

**Why cents?**
- ✅ No floating-point rounding errors
- ✅ Standard practice in payment systems (Stripe, PayPal, etc.)
- ✅ Integer arithmetic is exact and fast
- ✅ Easy conversion: `dollars = cents / 100`

**Example:**
```go
// In your application
coursePriceCents := 9999  // $99.99
orderTotal := course1Price + course2Price  // 9999 + 14999 = 24998 ($249.98)
```

### Query organization

Queries are organized by entity and operation type:
- **Read queries** (`*_read.sql`) - SELECT operations, includes pagination
- **Write queries** (`*_write.sql`) - INSERT, UPDATE, DELETE operations

### Adding a new SQL query

1. **Add query to appropriate file** in `query/<entity>/` directory:
   ```sql
   -- In query/users/users_read.sql
   -- name: GetUserByEmail :one
   SELECT * FROM users WHERE email = ?;
   ```

2. **Regenerate code**:
   ```bash
   sqlc generate
   ```

3. **Use in your code**:
   ```go
   user, err := queries.GetUserByEmail(ctx, "user@example.com")
   ```

### Pagination

All list queries support pagination with `LIMIT` and `OFFSET`:
```sql
-- name: ListUsers :many
SELECT id, email, name, age, created_at
FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?;
```

Generated code:
```go
users, err := queries.ListUsers(ctx, db.ListUsersParams{
    Limit:  10,
    Offset: 0,
})
```

### Query annotations

- `:one` - Returns a single row (error if not found)
- `:many` - Returns multiple rows (slice)
- `:exec` - Executes without returning data
- `:execrows` - Returns number of affected rows

### Nullable fields

Use `sqlc.narg()` for optional parameters:
```sql
-- name: UpdateUser :exec
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    age = COALESCE(sqlc.narg('age'), age)
WHERE id = ?;
```

### Defensive Programming

**LEFT JOIN for orphaned data:**
```sql
-- Won't fail if course is deleted
SELECT e.*, COALESCE(c.title, '') as course_title
FROM enrollments e
LEFT JOIN courses c ON e.course_id = c.id;
```

**COALESCE for non-nullable columns:**
```sql
-- Prevents NULL in required fields
COALESCE(c.slug, '') as course_slug,
COALESCE(c.title, '') as course_title
```

**Result:**
- ✅ Queries don't break if foreign key references are missing
- ✅ Shows enrollments with empty strings instead of hiding records
- ✅ Easy to identify and fix data integrity issues

## Best Practices

### When to use what:

**Use `json_group_array()`:**
- When returning parent entity with all children (one-to-many)
- Example: Get course with all enrolled students
- Benefit: Single query, no N+1 problem

**Use regular JOINs:**
- When filtering or paginating child records
- Example: List enrollments with pagination
- Benefit: More flexible filtering

**Use transactions:**
- When multiple database operations must succeed together
- Example: Create order + order items + enrollments
- Benefit: Atomicity, all or nothing

**Prices in cents:**
- Always store monetary values as integers in cents
- Example: $99.99 = 9999
- Benefit: No floating-point precision errors

## Docker & PostgreSQL

### Quick Start with PostgreSQL

```bash
# 1. Setup environment
make setup

# 2. Start PostgreSQL
make docker-up

# 3. Verify it's running
make docker-ps

# 4. Open database shell
make db-shell
```

### Docker Commands

```bash
make docker-up      # Start PostgreSQL in background
make docker-down    # Stop PostgreSQL
make docker-logs    # View PostgreSQL logs
make docker-clean   # Remove everything (⚠️ destroys data)
make db-shell       # Open psql shell
make db-reset       # Drop and recreate database
```

### Using PostgreSQL

After starting PostgreSQL with `make docker-up`, the application automatically detects and uses it based on the `DATABASE_URL` environment variable:

```bash
# In .env file
DATABASE_URL=postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable

# Or export directly
export DATABASE_URL="postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable"
./bin/gosql user-add -e "alice@example.com" -n "Alice"
```

### Switching Between Databases

The application automatically detects which database to use:

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

### Environment Configuration

```bash
# .env file (created by `make setup`)
POSTGRES_USER=gosql
POSTGRES_PASSWORD=dev_password_123
POSTGRES_DB=gosql_dev
POSTGRES_HOST=localhost
POSTGRES_PORT=5432

# To use PostgreSQL, add:
DATABASE_URL=postgres://gosql:dev_password_123@localhost:5432/gosql_dev?sslmode=disable
```

**Important:**
- `.env` is gitignored (contains secrets)
- `.env.example` is committed (template)
- Always use `.env.example` as reference

### Data Persistence

Data is stored in Docker volumes:
- `gosql_postgres_data` - Database files
- Persists across container restarts
- Only removed with `make docker-clean`

## Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design and architecture
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Development guide for contributors
- **[sqlc documentation](https://docs.sqlc.dev/)** - sqlc reference
- **[Kong CLI documentation](https://github.com/alecthomas/kong)** - CLI framework

## Database Migrations

> **TL;DR**: Migrations run automatically. Use `goose` CLI for manual operations. New tables: `lessons` and `course_reviews`.

### Migration System

This project uses [goose](https://github.com/pressly/goose) for database migrations with automatic migration on startup.

**Automatic Migrations:**
- Migrations run automatically when the application starts
- Separate migration files for SQLite and PostgreSQL
- Version tracked in `goose_db_version` table

**Migration Files:**
```
internal/migrate/migrations/
├── 001_initial_schema.sql        # Core tables (users, courses, orders, etc.)
├── 002_add_lessons_table.sql     # Course lessons
└── 003_add_reviews_table.sql     # Course reviews

internal/migrate/migrations/postgres/
├── 001_initial_schema.sql
├── 002_add_lessons_table.sql
└── 003_add_reviews_table.sql
```

### Manual Migration Commands

Install goose CLI (if not already installed):
```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.27.1
```

Run migrations manually:
```bash
# SQLite
cd internal/migrate
goose -dir migrations sqlite /path/to/data.db up

# PostgreSQL
goose -dir migrations/postgres postgres "user=gosql password=dev_password_123 dbname=gosql_dev sslmode=disable" up
```

Check migration status:
```bash
# SQLite
goose -dir migrations sqlite /path/to/data.db status

# PostgreSQL
goose -dir migrations/postgres postgres "connection_string" status
```

Rollback last migration:
```bash
# SQLite
goose -dir migrations sqlite /path/to/data.db down

# PostgreSQL
goose -dir migrations/postgres postgres "connection_string" down
```

### Creating New Migrations

```bash
cd internal/migrate

# Create SQLite migration
goose -dir migrations create add_new_table sql

# Create PostgreSQL migration (adapt SQL syntax)
goose -dir migrations/postgres create add_new_table sql
```

**Migration Template:**
```sql
-- +goose Up
CREATE TABLE example (
    id INTEGER PRIMARY KEY AUTOINCREMENT,  -- SQLite
    -- id SERIAL PRIMARY KEY,              -- PostgreSQL
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP  -- SQLite
    -- created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- PostgreSQL
);

-- +goose Down
DROP TABLE IF EXISTS example;
```

**Key Differences:**
| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| Auto-increment | `INTEGER PRIMARY KEY AUTOINCREMENT` | `SERIAL PRIMARY KEY` |
| Timestamp | `DATETIME` | `TIMESTAMP` |
| Foreign keys | Inline or separate | Inline with `REFERENCES` |

### New Tables (v3 Migrations)

**lessons** - Course lessons (1-to-many with courses)
- `id`, `course_id` (FK to courses), `title`, `description`, `created_at`
- Example: "Introduction to Go", "Variables and Types"

**course_reviews** - Course reviews and ratings
- `id`, `course_id` (FK to courses), `user_id` (FK to users), `rating` (1-5), `comment`, `created_at`
- Rating constraint: `CHECK (rating BETWEEN 1 AND 5)`
- Example: 5 stars, "Great course!"

## Database Schema

### Core Tables

**users** - User accounts
- `id`, `email` (UNIQUE), `name`, `age`, `created_at`

**courses** - Available courses
- `id`, `slug` (UNIQUE), `title`, `price` (in cents)

**enrollments** - User course access
- `id`, `user_id`, `course_id`, `enrolled_at`, `status`, `order_id`
- UNIQUE constraint on (user_id, course_id)

**orders** - Purchase records
- `id`, `user_id`, `total_amount` (in cents), `status`, `payment_method`
- Status: pending, completed, failed, refunded

**order_items** - Items in each order
- `id`, `order_id`, `course_id`, `price` (historical)
- UNIQUE constraint on (order_id, course_id)

### Relationships

```
users
  ├─→ orders (one-to-many)
  │     └─→ order_items (one-to-many)
  │           └─→ courses (many-to-one)
  └─→ enrollments (one-to-many)
        ├─→ courses (many-to-one)
        └─→ orders (many-to-one, optional)
```

### Indexes

All foreign keys are indexed for performance:
- `idx_orders_user_id`, `idx_orders_status`
- `idx_order_items_order_id`, `idx_order_items_course_id`
- `idx_enrollments_order_id`

## Testing

This project has comprehensive **unit tests** and **integration tests**.

### Quick Start:
```bash
make test              # Unit tests only
make test-integration  # Integration tests only
make test-all          # Both unit and integration tests
```

### Run tests with coverage:
```bash
make test-cover  # Generates coverage.html
```

### Run specific test suites:
```bash
make test-models    # Model conversion tests
make test-repo      # Repository/data access tests
make test-service   # Service/business logic tests
```

### Test Statistics:

**Unit Tests**: 44 tests
- **Models**: 23.0% coverage - Type conversion functions
- **Repository**: 53.7% coverage - Data access, transactions, error mapping
- **Service**: 34.6% coverage - Business logic, error handling

**Integration Tests**: 8 tests
- Real database files with transaction rollback
- Complete end-to-end user flows
- Build tag: `//go:build integration`

**Total**: 52 tests

### What's tested:

✅ **Model conversions** - `sql.Null*` to pointer types
✅ **CRUD operations** - Create, Read, Update, Delete for all entities
✅ **Error handling** - `ErrNotFound`, `ErrConflict` mapping
✅ **Transactions** - Order creation with enrollments (atomicity)
✅ **Transaction rollback** - Integration tests clean up automatically
✅ **Pagination** - List operations with LIMIT/OFFSET
✅ **Duplicate prevention** - UNIQUE constraints
✅ **Course ownership** - Verify purchase → enrollment flow (both paid and free)
✅ **Friendly error messages** - Service layer user-facing errors
✅ **Complete flows** - User signup → course purchase → access verification

### Test structure:

```
internal/
├── testutil/
│   ├── testutil.go           # Unit test helpers (in-memory DB)
│   └── integration.go        # Integration helpers (real DB + transactions)
├── models/
│   └── models_test.go        # 5 tests
├── repository/
│   └── repository_test.go    # 22 tests
├── service/
│   └── service_test.go       # 17 tests
└── tests/
    └── integration_test.go   # 8 integration tests
```

**Unit tests** use in-memory SQLite for speed.
**Integration tests** use real database files with automatic transaction rollback.

See [docs/TESTING.md](docs/TESTING.md) for detailed testing guide.

## TODO

- [x] Add MIT License file
- [x] Add unit tests for service layer
- [x] Add integration tests for repository layer
- [x] Add migration tracking system (goose with version tracking)
- [ ] Add proper versioning (semantic versioning)
- [ ] Add GitHub Actions CI/CD
  - [ ] Run tests
  - [ ] Check `sqlc generate` is up to date
  - [ ] Lint Go code
- [ ] Add order cancellation/refund with enrollment revocation
- [ ] Add course revenue analytics queries
- [ ] Add user purchase history export

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
