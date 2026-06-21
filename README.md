# Go SQL CLI Tool

A command-line interface for managing SQLite databases with support for courses and users.

## Features

- **Repository Pattern**: Clean separation of data access logic
- **JSON Output**: All commands output structured JSON
- **Nullable Fields**: Proper handling of optional database fields
- **Type Safety**: Strong typing with Go structs
- **Kong CLI**: Modern command-line parsing with auto-generated help

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

## Example JSON Output

```json
{
  "id": 1,
  "email": "john@example.com",
  "name": "John Doe",
  "age": 30,
  "created_at": "2026-06-21T09:58:05Z"
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
│       ├── models.go        # Data models (Course, User)
│       ├── course.go        # CourseRepository
│       └── user.go          # UserRepository
│
├── examples/
│   └── user_example.go      # Example usage
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
```

## Documentation

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed documentation.
