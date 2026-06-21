# Project Architecture

## Overview

This project is a CLI tool for managing SQLite databases, built with Go and using the Kong library for command-line parsing.

## Project Structure

```
.
├── cmd/app/
│   └── main.go              # CLI commands and application entry point
│
├── internal/
│   ├── database/
│   │   └── sqlite.go        # Database connection and schema management
│   │
│   └── storage/
│       ├── models.go        # Data models (Course, User)
│       ├── course.go        # CourseRepository - DB operations for courses
│       └── user.go          # UserRepository - DB operations for users
│
├── examples/
│   └── user_example.go      # Example usage of UserRepository
│
└── docs/
    └── ARCHITECTURE.md      # This file
```

## Layer Responsibilities

### 1. `cmd/app/` - CLI Layer
- Defines Kong CLI structure and commands
- Handles user input/output formatting
- Delegates business operations to repositories
- **No direct SQL queries**

### 2. `internal/database/` - Database Management
- Database connection setup
- Schema initialization
- Connection pool configuration
- Database-agnostic configuration

### 3. `internal/storage/` - Data Access Layer
- **Models**: Go structs representing database tables
- **Repositories**: Encapsulate all SQL queries for each entity
- Handle nullable fields using pointers (`*string`, `*int`)
- Return domain models, not raw SQL results

## Design Patterns

### Repository Pattern
Each entity (Course, User) has its own repository:

```go
type CourseRepository struct {
    db *sql.DB
}

func (r *CourseRepository) Create(ctx context.Context, course Course) (int64, error)
func (r *CourseRepository) List(ctx context.Context, limit, offset int, orderBy string) ([]Course, error)
func (r *CourseRepository) FindByIDs(ctx context.Context, ids []int64) ([]Course, error)
// ... etc
```

**Benefits:**
- Testable (easy to mock)
- Reusable across different commands
- Single source of truth for queries
- Type-safe operations

### Dependency Injection with Kong
Dependencies are injected into command `Run()` methods:

```go
func (cmd *CourseAddCmd) Run(ctx context.Context, repo *storage.CourseRepository) error {
    // Use repo to interact with database
}
```

Binding is done in `main()`:
```go
kong.BindTo(ctx, (*context.Context)(nil))  // For interfaces
kong.Bind(courseRepo)                       // For concrete types
```

## Handling Nullable Fields

### Problem
SQL allows NULL values, but Go doesn't have built-in nullable types.

### Solution
Use pointers for nullable fields:

```go
type User struct {
    ID        int64   `json:"id"`
    Email     string  `json:"email"`
    Name      *string `json:"name,omitempty"`  // Can be NULL
    Age       *int    `json:"age,omitempty"`   // Can be NULL
    CreatedAt string  `json:"created_at"`
}
```

### Working with Nullable Fields

**Creating users:**
```go
// With values
userRepo.Create(ctx, storage.CreateUserDTO{
    Email: "user@example.com",
    Name:  stringPtr("John Doe"),
    Age:   intPtr(30),
})

// With NULL
userRepo.Create(ctx, storage.CreateUserDTO{
    Email: "anonymous@example.com",
    Name:  nil,  // NULL in database
    Age:   nil,  // NULL in database
})
```

**Helper functions:**
```go
func stringPtr(s string) *string { return &s }
func intPtr(i int) *int { return &i }
```

**Reading nullable fields:**
```go
if user.Name != nil {
    fmt.Println("Name:", *user.Name)
} else {
    fmt.Println("Name: NULL")
}
```

### Partial Updates with COALESCE

The `Update` method uses SQL `COALESCE` to only update non-nil fields:

```sql
UPDATE users
SET name = COALESCE(?, name),  -- If ? is NULL, keep existing value
    age  = COALESCE(?, age)
WHERE id = ?
```

**Example:**
```go
// Update only name, keep existing age
userRepo.Update(ctx, storage.UpdateUserDTO{
    ID:   1,
    Name: stringPtr("New Name"),
    Age:  nil,  // Don't change age
})
```

## Database Schema

### Courses Table
```sql
CREATE TABLE courses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    price INTEGER NOT NULL DEFAULT 0
)
```

### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT,                              -- Nullable
    age INTEGER,                            -- Nullable
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
```

## Key Principles

1. **Separation of Concerns**: Each layer has a single responsibility
2. **Dependency Injection**: Dependencies flow from main → commands
3. **Type Safety**: Use Go types, not raw SQL
4. **Error Handling**: Wrap errors with context using `fmt.Errorf`
5. **Context Usage**: Pass context for cancellation and timeouts
6. **Nullable Fields**: Use pointers for optional database fields
