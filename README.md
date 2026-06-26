# Go SQL CLI Tool

A command-line interface for managing an **e-learning platform** with SQLite. Supports courses, users, enrollments, and a complete **order/payment system**. Built with **sqlc** for type-safe SQL queries and following **Clean Architecture** principles.

## Features

- **E-Commerce System**: Complete order management with purchase tracking and enrollments
- **sqlc**: Type-safe SQL queries generated from SQL files
- **Clean Architecture**: Service → Repository → Database layer separation
- **Domain Error Handling**: Custom errors (`ErrNotFound`, `ErrConflict`) with clear messages
- **JSON Aggregation**: Uses `json_group_array()` for efficient nested data queries
- **JSON Output**: All commands output structured JSON
- **Nullable Fields**: Proper handling of optional database fields with pointer types
- **Type Safety**: Strong typing with Go structs and sqlc-generated code
- **Kong CLI**: Modern command-line parsing with auto-generated help
- **Prepared Statements**: Bulk operations using prepared statements for performance
- **Upsert Support**: Insert or update (ON CONFLICT) for both courses and users
- **Transactions**: Atomic multi-step operations (order creation, enrollments)
- **Pagination**: All list operations support LIMIT/OFFSET
- **Defensive Queries**: LEFT JOIN queries handle orphaned data gracefully

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

## Documentation

- [docs/EXPORTING_PACKAGES.md](docs/EXPORTING_PACKAGES.md) - Guide for using as a library
- [examples/README_TRANSACTIONS.md](examples/README_TRANSACTIONS.md) - Detailed transaction guide
- [sqlc documentation](https://docs.sqlc.dev/) - sqlc reference
- [Kong CLI documentation](https://github.com/alecthomas/kong) - CLI framework reference

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

This project has comprehensive unit tests for all layers.

### Run all tests:
```bash
make test
```

### Run tests with coverage:
```bash
make test-cover
# Opens coverage.html in browser
```

### Run specific test suites:
```bash
make test-models    # Model conversion tests
make test-repo      # Repository/data access tests
make test-service   # Service/business logic tests
```

### Test coverage:
- **Models**: 23.0% - Type conversion functions
- **Repository**: 53.7% - Data access, transactions, error mapping
- **Service**: 34.6% - Business logic, error handling

### What's tested:

✅ **Model conversions** - `sql.Null*` to pointer types
✅ **CRUD operations** - Create, Read, Update, Delete for all entities
✅ **Error handling** - `ErrNotFound`, `ErrConflict` mapping
✅ **Transactions** - Order creation with enrollments (atomicity)
✅ **Pagination** - List operations with LIMIT/OFFSET
✅ **Duplicate prevention** - UNIQUE constraints
✅ **Course ownership** - Verify purchase → enrollment flow
✅ **Friendly error messages** - Service layer user-facing errors

### Test structure:

```
internal/
├── testutil/
│   └── testutil.go           # Test helpers (in-memory DB, seed data)
├── models/
│   └── models_test.go        # 5 tests
├── repository/
│   └── repository_test.go    # 22 tests
└── service/
    └── service_test.go       # 17 tests
```

All tests use **in-memory SQLite** for speed and isolation.

## TODO

- [x] Add MIT License file
- [x] Add unit tests for service layer
- [x] Add integration tests for repository layer
- [ ] Export packages for use as a library
  - [ ] Export `internal/service` as `pkg/service` (public API)
  - [ ] Export `internal/models` as `pkg/models` (public types)
  - [ ] Export `internal/repository` as `pkg/repository` (optional, advanced usage)
  - [ ] Keep `internal/db` internal (sqlc-generated code)
- [ ] Add proper versioning (semantic versioning)
- [ ] Add GitHub Actions CI/CD
  - [ ] Run tests
  - [ ] Check `sqlc generate` is up to date
  - [ ] Lint Go code
- [ ] Add example usage as a library (not just CLI)
- [ ] Document public API with godoc comments
- [ ] Add migration tracking system (instead of running all migrations every time)
- [ ] Add order cancellation/refund with enrollment revocation
- [ ] Add course revenue analytics queries
- [ ] Add user purchase history export

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
