# Go SQL CLI Tool

A command-line interface for managing SQLite databases with support for courses, users, and enrollments. Built with **sqlc** for type-safe SQL queries.

## Features

- **sqlc**: Type-safe SQL queries generated from SQL files
- **Repository Pattern**: Clean separation of data access logic
- **JSON Output**: All commands output structured JSON
- **Nullable Fields**: Proper handling of optional database fields
- **Type Safety**: Strong typing with Go structs and sqlc-generated code
- **Kong CLI**: Modern command-line parsing with auto-generated help
- **Prepared Statements**: Bulk operations using prepared statements for performance
- **Upsert Support**: Insert or update (ON CONFLICT) for both courses and users
- **Transactions**: Atomic operations with rollback support for data integrity

## Requirements

- Go 1.25.0 or higher
- **sqlc v1.31.1** - [Installation guide](https://docs.sqlc.dev/en/latest/overview/install.html)

```bash
# Install sqlc (macOS)
brew install sqlc

# Or download from GitHub releases
# https://github.com/sqlc-dev/sqlc/releases/tag/v1.31.1
```

## Installation

```bash
# Clone the repository
git clone <repo-url>
cd hexlet-go-sql

# Build the binary
make build
```

## Development

### Regenerating sqlc code

After modifying SQL schemas (`migrations/*.sql`) or queries (`query/*.sql`):

```bash
sqlc generate
```

This will regenerate the type-safe Go code in `internal/db/`.

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
│   │   └── database.go          # Database connection and schema initialization
│   │
│   ├── db/                      # sqlc-generated code (DO NOT EDIT MANUALLY)
│   │   ├── db.go                # Database interface and Queries struct
│   │   ├── models.go            # Auto-generated models from schema
│   │   ├── querier.go           # Interface for all queries
│   │   ├── courses.sql.go       # Generated course queries
│   │   ├── users.sql.go         # Generated user queries
│   │   └── enrollments.sql.go  # Generated enrollment queries
│   │
│   ├── models/
│   │   └── models.go            # JSON-friendly model converters
│   │
│   └── repository/
│       └── repository.go        # Transaction wrapper and business logic
│
├── migrations/
│   └── 001_schema.sql           # Database schema definition
│
├── query/
│   ├── courses/
│   │   ├── courses_read.sql     # Read queries (SELECT)
│   │   └── courses_write.sql    # Write queries (INSERT, UPDATE, DELETE)
│   ├── users/
│   │   ├── users_read.sql       # Read queries (SELECT)
│   │   └── users_write.sql      # Write queries (INSERT, UPDATE, DELETE)
│   └── enrollments/
│       ├── enrollments_read.sql # Read queries (SELECT)
│       └── enrollments_write.sql # Write queries (INSERT, UPDATE)
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

## Quick Reference

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

## Documentation

- [examples/README_TRANSACTIONS.md](examples/README_TRANSACTIONS.md) - Detailed transaction guide
- [sqlc documentation](https://docs.sqlc.dev/) - sqlc reference
- [Kong CLI documentation](https://github.com/alecthomas/kong) - CLI framework reference
