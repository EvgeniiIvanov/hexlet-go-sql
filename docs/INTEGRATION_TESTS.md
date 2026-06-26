# Integration Tests

## Overview

Integration tests verify complete user flows using real SQLite database files with automatic transaction rollback for test isolation.

## Key Design Decisions

### 1. Build Tag Separation

```go
//go:build integration
```

- Integration tests are **opt-in** via build tags
- Unit tests run fast without integration tests
- CI can run them separately for better feedback
- Run with: `go test -tags=integration`

### 2. Transaction Rollback Pattern

Instead of cleaning up data after tests, we use transaction rollback:

```go
testutil.WithTxContext(t, db, func(ctx context.Context, tx *sql.Tx) {
    // All database operations here
    // Automatic rollback after test completes
})
```

**Benefits:**
- ✅ Perfect isolation: No data pollution between tests
- ✅ No cleanup code needed
- ✅ Fast: Rollback is instant
- ✅ Safe: Even failed tests clean up

### 3. Repository Transaction Support

The repository supports **both** modes:

**Mode 1: Normal usage** (unit tests, production)
```go
repo := repository.New(db)
// Uses internal withTx() for transactions
```

**Mode 2: Existing transaction** (integration tests)
```go
queries := db.New(tx)
repo := repository.NewWithQueries(queries)
// Detects existing transaction, no nesting
```

Implementation in `withTx()`:
```go
func (r *Repository) withTx(ctx context.Context, fn func(*db.Queries) error) error {
    // If db is nil, we're already in a transaction
    if r.db == nil {
        return fn(r.queries)
    }
    // Otherwise create new transaction
    tx, _ := r.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    return fn(r.queries.WithTx(tx))
}
```

### 4. Migration Path Resolution

Integration tests can run from different directories, so we try multiple paths:

```go
possiblePaths := [][]string{
    {"migrations/001_schema.sql", "migrations/002_add_orders.sql"},
    {"../../migrations/001_schema.sql", "../../migrations/002_add_orders.sql"},
    {"../../../migrations/001_schema.sql", "../../../migrations/002_add_orders.sql"},
}
```

This makes tests robust to different execution contexts.

## Test Scenarios

### 1. User Creation and Retrieval
Tests basic CRUD with verification:
- Create user
- Retrieve by ID
- Verify data matches

### 2. Duplicate Email Error
Tests constraint enforcement:
- Create user with email
- Attempt duplicate: Should fail with friendly error

### 3. Course Purchase Flow
Tests the complete e-commerce flow:
1. Create user
2. Create course
3. Verify user doesn't own course initially
4. Purchase course (creates order + enrollment)
5. Verify user owns course after purchase
6. Verify enrollment links to order

### 4. Multiple Course Purchase
Tests bulk operations:
- Purchase 3 courses in single order
- Verify total amount = sum of course prices
- Verify all 3 enrollments created
- Verify user owns all 3 courses

### 5. Duplicate Purchase Prevention
Tests business rules:
- Purchase course once
- Attempt second purchase: Should fail

### 6. Order with Items Retrieval
Tests JSON aggregation:
- Create order with multiple items
- Retrieve with `GetOrderWithItems()`
- Verify JSON contains all items
- Verify price calculations

### 7. User Orders History
Tests listing with relationships:
- Create multiple orders for user
- Retrieve with `GetUserOrders()`
- Verify each order has correct items

### 8. Free Enrollment
Tests alternative enrollment path:
- Enroll user without creating order
- Verify `order_id` is NULL
- Verify user still owns course

## Running Integration Tests

```bash
# Run all integration tests
make test-integration

# Run with verbose output
make test-integration-v

# Run specific test
go test -tags=integration ./internal/tests/... -run TestCoursePurchase -v

# Run both unit and integration
make test-all
```

## Common Patterns

### Pattern 1: Basic Test Structure

```go
func TestSomething(t *testing.T) {
    db := testutil.SetupIntegrationDB(t)
    
    testutil.WithTxContext(t, db, func(ctx context.Context, tx *sql.Tx) {
        queries := db.New(tx)
        repo := repository.NewWithQueries(queries)
        svc := service.New(repo)
        
        // Your test code here
    })
}
```

### Pattern 2: Multi-Step Flow

```go
// Step 1: Setup
user, _ := svc.CreateUser(ctx, "test@example.com", ...)
course, _ := svc.CreateCourse(ctx, "test-course", "Test", 1000)

// Step 2: Verify initial state
owns, _ := svc.CheckUserOwnsCourse(ctx, user.ID, course.ID)
if owns {
    t.Error("should not own initially")
}

// Step 3: Perform action
order, _ := svc.CreateOrder(ctx, user.ID, []int64{course.ID}, "card")

// Step 4: Verify final state
owns, _ = svc.CheckUserOwnsCourse(ctx, user.ID, course.ID)
if !owns {
    t.Error("should own after purchase")
}
```

## Troubleshooting

### Test fails with "no such table"
- Migrations didn't run
- Check migration file paths in `integration.go`

### Test sees data from previous test
- Transaction rollback didn't work
- Check that you're using `WithTxContext`
- Verify repository uses `NewWithQueries(tx)` not `New(db)`

### "nil pointer dereference" in repository
- Repository trying to use `r.db.BeginTx()` when `r.db` is nil
- This means `withTx()` doesn't detect existing transaction
- Check that `NewWithQueries` is used correctly

## Best Practices

1. **One flow per test**: Don't test multiple unrelated flows in one test
2. **Clear test names**: Name describes the complete scenario
3. **Verify state changes**: Check before and after states
4. **Test error cases**: Not just happy paths
5. **Use subtests**: Group related assertions with `t.Run()`

## Future Enhancements

- Add performance benchmarks for order creation
- Add tests for concurrent order creation
- Add tests for refund flow (when implemented)
- Add tests with large datasets (1000+ courses)
