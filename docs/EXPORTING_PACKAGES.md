# Exporting Packages for Library Usage

This document outlines the strategy for exporting internal packages so this project can be used as a library, not just a CLI tool.

## Current Structure (CLI-only)

```
internal/
├── database/    # Database connection (internal)
├── db/          # sqlc-generated (internal)
├── models/      # Type conversions (should be public)
├── repository/  # Data access layer (could be public)
└── service/     # Business logic (should be public)
```

## Proposed Structure (Library + CLI)

```
pkg/
├── service/     # Public API - business logic
│   └── service.go
├── models/      # Public types - JSON-friendly models
│   └── models.go
└── repository/  # Optional - for advanced users
    ├── errors.go
    └── repository.go

internal/
├── database/    # Stay internal - implementation detail
└── db/          # Stay internal - sqlc-generated

cmd/app/         # CLI consumer of pkg/
```

## Migration Steps

### Step 1: Move models to pkg/models

```bash
mkdir -p pkg/models
cp internal/models/models.go pkg/models/
# Update package name from "models" to "models"
# Update imports in service and CLI
```

### Step 2: Move service to pkg/service

```bash
mkdir -p pkg/service
cp internal/service/service.go pkg/service/
# Update imports
```

### Step 3: Move repository to pkg/repository

```bash
mkdir -p pkg/repository
cp internal/repository/*.go pkg/repository/
# Update imports
```

### Step 4: Update all imports

```go
// OLD
import "example.com/go-sql/internal/models"
import "example.com/go-sql/internal/service"

// NEW
import "example.com/go-sql/pkg/models"
import "example.com/go-sql/pkg/service"
```

## Public API Example

Once exported, users could use it as a library:

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    
    "example.com/go-sql/pkg/service"
    "example.com/go-sql/pkg/models"
    "example.com/go-sql/pkg/repository"
    
    _ "modernc.org/sqlite"
)

func main() {
    // 1. Open database connection
    db, err := sql.Open("sqlite", "myapp.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // 2. Create repository
    repo := repository.New(db)
    
    // 3. Create service (main API)
    svc := service.New(repo)
    
    // 4. Use the service
    ctx := context.Background()
    
    // Create a user
    user, err := svc.CreateUser(ctx, "alice@example.com", nil, nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created user: %+v\n", user)
    
    // Create a course
    course, err := svc.CreateCourse(ctx, "go-101", "Go Programming", 9999)
    if err != nil {
        panic(err)
    }
    
    // Purchase course (creates order + enrollment)
    order, err := svc.CreateOrder(ctx, user.ID, []int64{course.ID}, "card")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Order created: %+v\n", order)
    
    // Check access
    hasAccess, err := svc.CheckUserOwnsCourse(ctx, user.ID, course.ID)
    if err != nil {
        panic(err)
    }
    fmt.Printf("User has access: %v\n", hasAccess)
}
```

## API Surface

### pkg/service (Main Public API)

```go
// Course operations
CreateCourse(ctx, slug, title string, price int64) (models.Course, error)
GetCourse(ctx, id int64) (models.Course, error)
ListCourses(ctx, limit, offset int64) ([]models.Course, error)
GetCourseWithEnrollments(ctx, id int64) (models.CourseWithEnrollments, error)

// User operations
CreateUser(ctx, email string, name sql.NullString, age sql.NullInt64) (models.User, error)
GetUser(ctx, id int64) (models.User, error)
ListUsers(ctx, limit, offset int64) ([]models.User, error)

// Order operations (E-commerce)
CreateOrder(ctx, userID int64, courseIDs []int64, paymentMethod string) (models.Order, error)
GetOrderWithItems(ctx, id int64) (models.OrderWithItems, error)
GetUserOrders(ctx, userID int64) ([]models.OrderWithItems, error)
CheckUserOwnsCourse(ctx, userID, courseID int64) (bool, error)
RefundOrder(ctx, orderID int64) error

// Enrollment operations
EnrollUser(ctx, userID, courseID int64) (models.Enrollment, error)
ListEnrollments(ctx, limit, offset int64) ([]models.EnrollmentWithDetails, error)
```

### pkg/repository (Advanced API)

For users who need direct data access or custom transactions:

```go
// Errors
var (
    ErrNotFound = errors.New("resource not found")
    ErrConflict = errors.New("resource already exists")
)

// Repository methods
New(db *sql.DB) *Repository
CreateOrderWithEnrollments(ctx, userID, courseIDs, paymentMethod) (Order, error)
// ... all other repository methods
```

### pkg/models (Public Types)

Clean JSON-friendly types:

```go
type User struct {
    ID        int64     `json:"id"`
    Email     string    `json:"email"`
    Name      *string   `json:"name,omitempty"`
    Age       *int64    `json:"age,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

type Order struct {
    ID            int64   `json:"id"`
    UserID        int64   `json:"user_id"`
    TotalAmount   int64   `json:"total_amount"`  // in cents
    Status        string  `json:"status"`
    PaymentMethod *string `json:"payment_method,omitempty"`
    // ...
}
```

## Benefits of Exporting

1. **Reusability**: Use the business logic in other projects (web app, mobile backend, etc.)
2. **Testing**: External packages can write tests against the public API
3. **Documentation**: godoc will generate nice API docs
4. **Versioning**: Can use semantic versioning for the library
5. **Separation**: Clear boundary between public API and implementation details

## Versioning Strategy

Once packages are exported, use semantic versioning:

```
v1.0.0 - Initial stable release
v1.1.0 - Add new features (backward compatible)
v1.1.1 - Bug fixes
v2.0.0 - Breaking changes (change service API, etc.)
```

## Next Steps

1. ✅ Add LICENSE file
2. [ ] Move `internal/models` → `pkg/models`
3. [ ] Move `internal/service` → `pkg/service`
4. [ ] Move `internal/repository` → `pkg/repository`
5. [ ] Update all imports
6. [ ] Add godoc comments to all exported functions
7. [ ] Add example usage in `examples/library_usage.go`
8. [ ] Add tests for public API
9. [ ] Tag v1.0.0 release
10. [ ] Publish to pkg.go.dev
