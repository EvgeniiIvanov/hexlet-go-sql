# Go SQL CLI Tool

A command-line interface for managing SQLite databases with support for courses and users.

## Features

- **Repository Pattern**: Clean separation of data access logic
- **JSON Output**: All commands output structured JSON
- **Nullable Fields**: Proper handling of optional database fields
- **Type Safety**: Strong typing with Go structs
- **Kong CLI**: Modern command-line parsing with auto-generated help
- **Prepared Statements**: Bulk operations using prepared statements for performance
- **Upsert Support**: Insert or update (ON CONFLICT) for both courses and users
- **Transactions**: Atomic operations with rollback support for data integrity

## Installation

```bash
make build
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
./bin/gosql course-list -l 5 -o 10 -r price
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
│   └── main.go              # CLI commands and entry point
│
├── internal/
│   ├── database/
│   │   └── sqlite.go        # Database connection and schema
│   │
│   └── storage/
│       ├── models.go        # Data models (Course, User, Enrollment)
│       ├── course.go        # CourseRepository
│       ├── user.go          # UserRepository
│       └── enrollment.go    # EnrollmentRepository with transactions
│
├── examples/
│   ├── enrollment_demo.sh   # Transaction demo script
│   └── bulk_performance_demo.sh # Bulk operations demo
│
└── docs/
    └── ARCHITECTURE.md      # Detailed architecture documentation
```

## Key Concepts

### Nullable Fields

Fields that can be NULL in the database use pointers in Go:

```go
type User struct {
    Name *string `json:"name,omitempty"`  // Can be NULL
    Age  *int    `json:"age,omitempty"`   // Can be NULL
}
```

### Repository Pattern

Each entity has its own repository:

```go
courseRepo.Create(ctx, CreateCourseDTO{...})
userRepo.List(ctx)
enrollmentRepo.EnrollUser(ctx, userID, courseID)
```

### Transactions

The `withTx` helper wraps operations in a database transaction:

```go
func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()  // Rollback if not committed

    if err := fn(tx); err != nil {
        return err  // Automatic rollback
    }

    return tx.Commit()  // Commit if no errors
}
```

Used in `EnrollUser` to ensure atomicity:
1. Check user exists
2. Check course exists
3. Create enrollment
4. If ANY step fails, rollback entire transaction

## Documentation

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed documentation.
