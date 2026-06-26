//go:build integration

package tests

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"example.com/go-sql/internal/db"
	"example.com/go-sql/internal/repository"
	"example.com/go-sql/internal/service"
	"example.com/go-sql/internal/testutil"
)

// TestUserCreationAndRetrieval tests the full flow of creating and retrieving a user
func TestUserCreationAndRetrieval(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		// Create repository and service with transaction
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		// Create a user
		user, err := svc.CreateUser(ctx,
			"john@example.com",
			sql.NullString{String: "John Doe", Valid: true},
			sql.NullInt64{Int64: 28, Valid: true})

		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Verify user was created with correct data
		if user.Email != "john@example.com" {
			t.Errorf("expected email john@example.com, got %s", user.Email)
		}
		if user.Name == nil || *user.Name != "John Doe" {
			t.Errorf("expected name 'John Doe', got %v", user.Name)
		}

		// Retrieve the user by ID
		retrievedUser, err := svc.GetUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to retrieve user: %v", err)
		}

		// Verify retrieved user matches created user
		if retrievedUser.ID != user.ID {
			t.Errorf("expected ID %d, got %d", user.ID, retrievedUser.ID)
		}
		if retrievedUser.Email != user.Email {
			t.Errorf("expected email %s, got %s", user.Email, retrievedUser.Email)
		}
	})
}

// TestDuplicateEmailError verifies that creating a user with duplicate email fails
func TestDuplicateEmailError(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		email := "duplicate@example.com"

		// Create first user
		_, err := svc.CreateUser(ctx, email,
			sql.NullString{String: "First User", Valid: true},
			sql.NullInt64{Valid: false})
		if err != nil {
			t.Fatalf("first user creation should succeed: %v", err)
		}

		// Attempt to create second user with same email
		_, err = svc.CreateUser(ctx, email,
			sql.NullString{String: "Second User", Valid: true},
			sql.NullInt64{Valid: false})

		// Should fail with conflict error
		if err == nil {
			t.Fatal("expected error for duplicate email, got nil")
		}

		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' error, got: %v", err)
		}
		if !strings.Contains(err.Error(), email) {
			t.Errorf("expected error to mention email %s, got: %v", email, err)
		}
	})
}

// TestCoursePurchaseFlow tests the complete e-commerce flow:
// 1. Create user
// 2. Create course
// 3. User purchases course (creates order + enrolls)
// 4. Verify user has access to course
func TestCoursePurchaseFlow(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		// Step 1: Create a user
		user, err := svc.CreateUser(ctx,
			"buyer@example.com",
			sql.NullString{String: "Test Buyer", Valid: true},
			sql.NullInt64{Int64: 30, Valid: true})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Step 2: Create a course
		course, err := svc.CreateCourse(ctx, "docker-101", "Docker Fundamentals", 12999)
		if err != nil {
			t.Fatalf("failed to create course: %v", err)
		}

		// Step 3: Verify user doesn't own the course yet
		ownsBefore, err := svc.CheckUserOwnsCourse(ctx, user.ID, course.ID)
		if err != nil {
			t.Fatalf("failed to check ownership: %v", err)
		}
		if ownsBefore {
			t.Error("user should not own course before purchase")
		}

		// Step 4: User purchases the course
		order, err := svc.CreateOrder(ctx, user.ID, []int64{course.ID}, "card")
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		// Step 5: Verify order was created correctly
		if order.Status != "completed" {
			t.Errorf("expected status 'completed', got %s", order.Status)
		}
		if order.TotalAmount != course.Price {
			t.Errorf("expected total %d, got %d", course.Price, order.TotalAmount)
		}

		// Step 6: Verify user now owns the course
		ownsAfter, err := svc.CheckUserOwnsCourse(ctx, user.ID, course.ID)
		if err != nil {
			t.Fatalf("failed to check ownership after purchase: %v", err)
		}
		if !ownsAfter {
			t.Error("user should own course after purchase")
		}

		// Step 7: Verify enrollment was created with correct order_id
		enrollments, err := svc.ListEnrollmentsByUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to list enrollments: %v", err)
		}
		if len(enrollments) != 1 {
			t.Fatalf("expected 1 enrollment, got %d", len(enrollments))
		}
		if enrollments[0].CourseID != course.ID {
			t.Errorf("expected course ID %d, got %d", course.ID, enrollments[0].CourseID)
		}
	})
}

// TestMultipleCoursePurchase tests purchasing multiple courses in a single order
func TestMultipleCoursePurchase(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		// Create user
		user, err := svc.CreateUser(ctx, "multi@example.com",
			sql.NullString{String: "Multi Buyer", Valid: true},
			sql.NullInt64{Valid: false})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Create multiple courses
		course1, err := svc.CreateCourse(ctx, "course-1", "Course One", 5000)
		if err != nil {
			t.Fatalf("failed to create course 1: %v", err)
		}

		course2, err := svc.CreateCourse(ctx, "course-2", "Course Two", 7500)
		if err != nil {
			t.Fatalf("failed to create course 2: %v", err)
		}

		course3, err := svc.CreateCourse(ctx, "course-3", "Course Three", 10000)
		if err != nil {
			t.Fatalf("failed to create course 3: %v", err)
		}

		// Purchase all three courses in one order
		order, err := svc.CreateOrder(ctx, user.ID,
			[]int64{course1.ID, course2.ID, course3.ID}, "paypal")
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		// Verify total amount (5000 + 7500 + 10000 = 22500)
		expectedTotal := int64(22500)
		if order.TotalAmount != expectedTotal {
			t.Errorf("expected total %d, got %d", expectedTotal, order.TotalAmount)
		}

		// Verify all enrollments were created
		enrollments, err := svc.ListEnrollmentsByUser(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to list enrollments: %v", err)
		}
		if len(enrollments) != 3 {
			t.Fatalf("expected 3 enrollments, got %d", len(enrollments))
		}

		// Verify user owns all three courses
		owns1, _ := svc.CheckUserOwnsCourse(ctx, user.ID, course1.ID)
		owns2, _ := svc.CheckUserOwnsCourse(ctx, user.ID, course2.ID)
		owns3, _ := svc.CheckUserOwnsCourse(ctx, user.ID, course3.ID)

		if !owns1 || !owns2 || !owns3 {
			t.Error("user should own all three purchased courses")
		}
	})
}

// TestDuplicateCoursePurchasePrevention verifies that buying the same course twice fails
func TestDuplicateCoursePurchasePrevention(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		// Setup: Create user and course
		user, _ := svc.CreateUser(ctx, "test@example.com",
			sql.NullString{Valid: false}, sql.NullInt64{Valid: false})
		course, _ := svc.CreateCourse(ctx, "test-course", "Test Course", 9999)

		// First purchase succeeds
		_, err := svc.CreateOrder(ctx, user.ID, []int64{course.ID}, "card")
		if err != nil {
			t.Fatalf("first purchase should succeed: %v", err)
		}

		// Second purchase should fail
		_, err = svc.CreateOrder(ctx, user.ID, []int64{course.ID}, "card")
		if err == nil {
			t.Fatal("expected error for duplicate purchase, got nil")
		}

		if !strings.Contains(err.Error(), "already enrolled") {
			t.Errorf("expected 'already enrolled' error, got: %v", err)
		}
	})
}

// TestOrderWithItemsRetrieval tests getting an order with its items via JSON aggregation
func TestOrderWithItemsRetrieval(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		// Create user and courses
		user, _ := svc.CreateUser(ctx, "order@example.com",
			sql.NullString{Valid: false}, sql.NullInt64{Valid: false})
		course1, _ := svc.CreateCourse(ctx, "item-1", "Item One", 3000)
		course2, _ := svc.CreateCourse(ctx, "item-2", "Item Two", 4000)

		// Create order
		order, err := svc.CreateOrder(ctx, user.ID, []int64{course1.ID, course2.ID}, "card")
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		// Retrieve order with items
		orderWithItems, err := svc.GetOrderWithItems(ctx, order.ID)
		if err != nil {
			t.Fatalf("failed to get order with items: %v", err)
		}

		// Verify order details
		if orderWithItems.ID != order.ID {
			t.Errorf("expected order ID %d, got %d", order.ID, orderWithItems.ID)
		}

		// Verify items were included
		if len(orderWithItems.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(orderWithItems.Items))
		}

		// Verify item prices
		totalFromItems := orderWithItems.Items[0].Price + orderWithItems.Items[1].Price
		if totalFromItems != orderWithItems.TotalAmount {
			t.Errorf("sum of item prices (%d) doesn't match order total (%d)",
				totalFromItems, orderWithItems.TotalAmount)
		}
	})
}

// TestUserOrdersHistory verifies getting all orders for a user
func TestUserOrdersHistory(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		// Create user and courses
		user, _ := svc.CreateUser(ctx, "history@example.com",
			sql.NullString{Valid: false}, sql.NullInt64{Valid: false})
		course1, _ := svc.CreateCourse(ctx, "hist-1", "Course 1", 1000)
		course2, _ := svc.CreateCourse(ctx, "hist-2", "Course 2", 2000)
		course3, _ := svc.CreateCourse(ctx, "hist-3", "Course 3", 3000)

		// Create multiple orders
		_, _ = svc.CreateOrder(ctx, user.ID, []int64{course1.ID}, "card")
		_, _ = svc.CreateOrder(ctx, user.ID, []int64{course2.ID, course3.ID}, "paypal")

		// Get order history
		orders, err := svc.GetUserOrders(ctx, user.ID)
		if err != nil {
			t.Fatalf("failed to get user orders: %v", err)
		}

		if len(orders) != 2 {
			t.Fatalf("expected 2 orders, got %d", len(orders))
		}

		// Verify first order has 1 item
		if len(orders[0].Items) != 1 {
			t.Errorf("expected first order to have 1 item, got %d", len(orders[0].Items))
		}

		// Verify second order has 2 items
		if len(orders[1].Items) != 2 {
			t.Errorf("expected second order to have 2 items, got %d", len(orders[1].Items))
		}
	})
}

// TestFreeEnrollmentWithoutOrder tests enrolling a user without creating an order
func TestFreeEnrollmentWithoutOrder(t *testing.T) {
	database := testutil.SetupIntegrationDB(t)

	testutil.WithTxContext(t, database, func(ctx context.Context, tx *sql.Tx) {
		queries := db.New(tx)
		repo := repository.NewWithQueries(queries)
		svc := service.New(repo)

		// Create user and course
		user, _ := svc.CreateUser(ctx, "free@example.com",
			sql.NullString{Valid: false}, sql.NullInt64{Valid: false})
		course, _ := svc.CreateCourse(ctx, "free-course", "Free Course", 0)

		// Enroll user directly (free enrollment)
		enrollment, err := svc.EnrollUser(ctx, user.ID, course.ID)
		if err != nil {
			t.Fatalf("failed to enroll user: %v", err)
		}

		// Verify enrollment has no order_id
		if enrollment.OrderID != nil {
			t.Errorf("expected nil order_id for free enrollment, got %v", *enrollment.OrderID)
		}

		// Verify user owns the course
		owns, err := svc.CheckUserOwnsCourse(ctx, user.ID, course.ID)
		if err != nil {
			t.Fatalf("failed to check ownership: %v", err)
		}
		if !owns {
			t.Error("user should own course after free enrollment")
		}
	})
}
