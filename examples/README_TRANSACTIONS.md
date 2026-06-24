# Transaction Implementation Guide

## Overview

This project demonstrates how to use database transactions in Go to ensure data integrity when performing complex operations that involve multiple database checks and inserts.

## The Transaction Helper

Located in `internal/storage/enrollment.go`:

```go
func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()  // Always rollback if we haven't committed

    if err := fn(tx); err != nil {
        return err  // Rollback happens automatically via defer
    }

    return tx.Commit()  // Only commit if no errors occurred
}
```

## How It Works

### 1. Begin Transaction
```go
tx, err := db.BeginTx(ctx, nil)
```
Creates a new transaction with default isolation level.

### 2. Defer Rollback
```go
defer tx.Rollback()
```
Ensures rollback happens if we exit early (error or panic). If `Commit()` is called, subsequent `Rollback()` calls are no-ops.

### 3. Execute Operations
```go
if err := fn(tx); err != nil {
    return err  // Rollback via defer
}
```
Executes the provided function with the transaction. Any error triggers rollback.

### 4. Commit
```go
return tx.Commit()
```
Only reached if no errors occurred. Makes all changes permanent.

## Use Case: EnrollUser

The `EnrollUser` method demonstrates a practical transaction use case:

```go
func (r *EnrollmentRepository) EnrollUser(ctx context.Context, userID, courseID int64) (*Enrollment, error) {
    var enrollment Enrollment

    err := withTx(ctx, r.db, func(tx *sql.Tx) error {
        // Step 1: Verify user exists
        var userExists bool
        err := tx.QueryRowContext(ctx, "SELECT 1 FROM users WHERE id = ?", userID).Scan(&userExists)
        if err == sql.ErrNoRows {
            return fmt.Errorf("user with id %d not found", userID)
        }

        // Step 2: Verify course exists
        var courseExists bool
        err = tx.QueryRowContext(ctx, "SELECT 1 FROM courses WHERE id = ?", courseID).Scan(&courseExists)
        if err == sql.ErrNoRows {
            return fmt.Errorf("course with id %d not found", courseID)
        }

        // Step 3: Create enrollment
        err = tx.QueryRowContext(ctx, `
            INSERT INTO enrollments(user_id, course_id, status)
            VALUES(?, ?, 'active')
            RETURNING id, user_id, course_id, enrolled_at, status
        `, userID, courseID).Scan(&enrollment.ID, &enrollment.UserID, ...)
        
        return err
    })

    return &enrollment, err
}
```

## Transaction Scenarios

### ✅ Success Case
All checks pass:
1. User exists ✓
2. Course exists ✓
3. Enrollment created ✓
4. **Transaction commits** → Enrollment saved to database

### ❌ User Not Found
1. User exists ✗ → Error returned
2. **Transaction rolls back** → Nothing saved
3. Database state unchanged

### ❌ Course Not Found
1. User exists ✓
2. Course exists ✗ → Error returned
3. **Transaction rolls back** → Nothing saved
4. Database state unchanged

### ❌ Duplicate Enrollment
1. User exists ✓
2. Course exists ✓
3. Enrollment creation fails (UNIQUE constraint) ✗
4. **Transaction rolls back** → Nothing saved
5. Original enrollment remains unchanged

## Benefits

### Atomicity
All operations succeed or none do. No partial state.

### Consistency
Database constraints (UNIQUE, FOREIGN KEY) are enforced within the transaction.

### Isolation
Other connections see either the old state or the new state, never a partial update.

### Durability
Once committed, changes are permanent (even if the process crashes immediately after).

## Testing Transactions

Run the demo script:
```bash
./examples/enrollment_demo.sh
```

This script tests:
- ✅ Successful enrollments
- ❌ Non-existent user (should rollback)
- ❌ Non-existent course (should rollback)
- ❌ Duplicate enrollment (UNIQUE constraint + rollback)

## CLI Examples

**Successful enrollment:**
```bash
$ ./bin/gosql enrollment-create -u 1 -c 1
{
  "id": 1,
  "user_id": 1,
  "course_id": 1,
  "enrolled_at": "2026-06-24T10:51:20Z",
  "status": "active"
}
```

**Failed enrollment (user not found):**
```bash
$ ./bin/gosql enrollment-create -u 999 -c 1
2026/06/24 14:51:30 run command: enroll user: user with id 999 not found
```

**Failed enrollment (duplicate):**
```bash
$ ./bin/gosql enrollment-create -u 1 -c 1
2026/06/24 14:51:35 run command: enroll user: create enrollment: 
constraint failed: UNIQUE constraint failed: enrollments.user_id, 
enrollments.course_id (2067)
```

## Key Takeaways

1. **Always use transactions** for operations with multiple related steps
2. **defer tx.Rollback()** is safe even after commit (becomes no-op)
3. **Check foreign keys** within the transaction for better error messages
4. **Let the database enforce constraints** (UNIQUE, FOREIGN KEY)
5. **Return early on errors** to trigger automatic rollback

## Related Files

- `internal/storage/enrollment.go` - Transaction implementation
- `internal/database/sqlite.go` - Schema with FOREIGN KEY constraints
- `cmd/app/main.go` - CLI commands for enrollment
- `examples/enrollment_demo.sh` - Interactive demonstration
