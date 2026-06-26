# Testing Guide

This document describes the testing strategy and how to run tests.

## Test Structure

```
internal/
├── testutil/
│   ├── testutil.go           # Unit test utilities (in-memory DB)
│   └── integration.go        # Integration test utilities (real DB + transactions)
├── models/
│   └── models_test.go        # Model conversion tests (5 tests)
├── repository/
│   └── repository_test.go    # Repository layer tests (22 tests)
├── service/
│   └── service_test.go       # Service layer tests (17 tests)
└── tests/
    └── integration_test.go   # Integration tests (8 tests)

Total: 44 unit tests + 8 integration tests = 52 tests
```

## Running Tests

### Quick Commands

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run ALL tests (unit + integration)
make test-all

# Run with verbose output
make test-verbose

# Run with coverage report
make test-cover  # Generates coverage.html

# Run specific test suites
make test-models
make test-repo
make test-service
```

### Coverage Results

```
Models:      23.0% coverage
Repository:  53.7% coverage
Service:     34.6% coverage
```

## Test Categories

### 1. Model Tests (`internal/models/models_test.go`)

Tests the conversion layer between sqlc types and JSON-friendly types.

**What's tested:**
- `FromDBUser` - Convert `sql.NullString` → `*string`
- `FromDBCourse` - Convert database course to model
- `FromDBOrder` - Convert database order with payment info
- Null field handling
- Array conversions

**Example:**
```go
func TestFromDBUser(t *testing.T) {
    t.Run("with null fields", func(t *testing.T) {
        dbUser := db.User{
            Name: sql.NullString{Valid: false},
        }
        user := FromDBUser(dbUser)
        
        if user.Name != nil {
            t.Error("expected nil name")
        }
    })
}
```

### 2. Repository Tests (`internal/repository/repository_test.go`)

Tests the data access layer with real SQLite database (in-memory).

**What's tested:**
- CRUD operations for users, courses, enrollments
- Error mapping (`sql.ErrNoRows` → `ErrNotFound`)
- UNIQUE constraint violations → `ErrConflict`
- Pagination
- **Transactions** (critical: order creation)
- Course ownership checks

**Key tests:**
```go
TestRepository_CreateOrderWithEnrollments
  ✓ Single course purchase
  ✓ Multiple courses in one order
  ✓ Transaction rollback on non-existent user
  ✓ Transaction rollback on non-existent course
  ✓ Duplicate purchase prevention
```

**Example:**
```go
func TestRepository_CreateOrderWithEnrollments(t *testing.T) {
    t.Run("create order with multiple courses", func(t *testing.T) {
        order, err := repo.CreateOrderWithEnrollments(
            ctx, userID, []int64{courseID1, courseID2}, "card")
        
        // Verify total amount
        if order.TotalAmount != 24998 {
            t.Errorf("expected 24998, got %d", order.TotalAmount)
        }
        
        // Verify enrollments were created
        owns1, _ := repo.CheckUserOwnsCourse(ctx, userID, courseID1)
        owns2, _ := repo.CheckUserOwnsCourse(ctx, userID, courseID2)
        if !owns1 || !owns2 {
            t.Error("user should own both courses")
        }
    })
}
```

### 3. Service Tests (`internal/service/service_test.go`)

Tests business logic and user-facing error messages.

**What's tested:**
- User-friendly error messages
- Model conversion integration
- Business rules
- Order creation workflow
- Ownership verification

**Key tests:**
```go
TestService_CreateOrder
  ✓ Order with single course
  ✓ Order with multiple courses
  ✓ Friendly error for non-existent user
  ✓ Friendly error for duplicate purchase
```

**Example:**
```go
func TestService_CreateOrder(t *testing.T) {
    t.Run("friendly error for duplicate purchase", func(t *testing.T) {
        _, err := svc.CreateOrder(ctx, userID, courseIDs, "card")
        
        if !strings.Contains(err.Error(), "already enrolled") {
            t.Errorf("expected 'already enrolled' in error, got: %v", err)
        }
    })
}
```

## Test Helpers

### SetupTestDB

Creates an in-memory SQLite database with migrations applied:

```go
func TestSomething(t *testing.T) {
    db := testutil.SetupTestDB(t)
    // db is ready to use, migrations applied
    // Automatically cleaned up after test
}
```

### SetupTestDBWithData

Creates database and seeds with common test data:

```go
func TestSomething(t *testing.T) {
    db, data := testutil.SetupTestDBWithData(t)
    
    // data.UserID1, data.UserID2 available
    // data.CourseID1, data.CourseID2 available
}
```

**Seeded data:**
- 2 users: alice@test.com, bob@test.com
- 2 courses: go-101 ($99.99), rust-101 ($149.99)

## Writing New Tests

### 1. Model Test Pattern

```go
func TestFromDBSomething(t *testing.T) {
    dbObj := db.Something{
        Field: sql.NullString{String: "value", Valid: true},
    }
    
    result := FromDBSomething(dbObj)
    
    if result.Field == nil || *result.Field != "value" {
        t.Errorf("expected 'value', got %v", result.Field)
    }
}
```

### 2. Repository Test Pattern

```go
func TestRepository_SomeOperation(t *testing.T) {
    db := testutil.SetupTestDB(t)
    repo := repository.New(db)
    ctx := context.Background()
    
    result, err := repo.SomeOperation(ctx, params)
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // assertions...
}
```

### 3. Service Test Pattern

```go
func TestService_SomeOperation(t *testing.T) {
    db, data := testutil.SetupTestDBWithData(t)
    repo := repository.New(db)
    svc := service.New(repo)
    ctx := context.Background()
    
    result, err := svc.SomeOperation(ctx, params)
    
    // Test both success and error cases
    // Verify user-friendly error messages
}
```

## Integration Tests

Integration tests verify the complete flow of operations using real database files and transactions.

### Key Features

1. **Build Tag**: Uses `//go:build integration` to separate from unit tests
2. **Real Database**: Creates temporary SQLite database files (not in-memory)
3. **Transaction Rollback**: Each test runs in a transaction that gets rolled back
4. **Complete Flows**: Tests entire user journeys, not just isolated methods

### Test Scenarios

```
TestUserCreationAndRetrieval         - Create user → Retrieve by ID
TestDuplicateEmailError              - Verify constraint enforcement
TestCoursePurchaseFlow               - User buys course → Gets access
TestMultipleCoursePurchase           - Buy multiple courses in one order
TestDuplicateCoursePurchasePrevention - Can't buy same course twice
TestOrderWithItemsRetrieval          - JSON aggregation works correctly
TestUserOrdersHistory                - Get all orders for a user
TestFreeEnrollmentWithoutOrder       - Enroll without payment
```

### Using WithTxContext Helper

```go
func TestSomething(t *testing.T) {
    db := testutil.SetupIntegrationDB(t)

    testutil.WithTxContext(t, db, func(ctx context.Context, tx *sql.Tx) {
        // Create repository with transaction
        queries := db.New(tx)
        repo := repository.NewWithQueries(queries)
        svc := service.New(repo)

        // Your test code here
        // All changes are rolled back after test
    })
}
```

### Benefits of Transaction Rollback

✅ **Test Isolation**: No data pollution between tests
✅ **No Cleanup Needed**: Rollback handles cleanup automatically
✅ **Safe Concurrent Testing**: Tests don't interfere with each other
✅ **Fast**: No need to truncate tables or reset sequences

### How It Works

The `WithTxContext` helper:
1. Begins a transaction
2. Passes the transaction to your test function
3. **Always rolls back** the transaction (using defer)
4. This ensures the database file stays clean

The repository's `NewWithQueries` constructor:
- Accepts sqlc queries bound to a transaction
- Detects when `db` is nil (meaning we're in a transaction)
- Skips creating nested transactions

### When to Use Integration Tests

Use integration tests for:
- **Multi-table operations** (orders + items + enrollments)
- **Transaction atomicity** verification
- **Constraint enforcement** (UNIQUE, FOREIGN KEY)
- **Complete user flows** (signup → purchase → access)
- **JSON aggregation queries** (complex JOINs)

Use unit tests for:
- Business logic validation
- Error message formatting
- Model conversions
- Edge cases with specific inputs

## Best Practices

1. **Use subtests**: Group related test cases with `t.Run()`
2. **Test error paths**: Don't just test happy paths
3. **Clear test names**: `test_description_expected_behavior`
4. **Meaningful assertions**: Error messages should explain what's wrong
5. **Isolate tests**: Each test should be independent
6. **Use helpers**: DRY principle with `testutil` package
7. **Tag appropriately**: Use build tags to separate integration tests

## CI/CD Integration

Add to GitHub Actions:

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run unit tests
        run: make test-cover
      - name: Run integration tests
        run: make test-integration
      - uses: codecov/codecov-action@v3
```

## Future Improvements

- [x] Add integration tests with transaction rollback
- [x] Separate unit and integration tests with build tags
- [ ] Increase coverage to 80%+
- [ ] Add end-to-end tests for CLI commands
- [ ] Add benchmark tests for bulk operations
- [ ] Add table-driven tests for complex scenarios
- [ ] Add mutation testing
- [ ] Add property-based testing for edge cases
