package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"example.com/go-sql/internal/repository"
	"example.com/go-sql/internal/testutil"
)

func TestService_CreateUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("create user successfully", func(t *testing.T) {
		user, err := svc.CreateUser(ctx, "test@example.com",
			sql.NullString{String: "Test User", Valid: true},
			sql.NullInt64{Int64: 25, Valid: true})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", user.Email)
		}
		if user.Name == nil || *user.Name != "Test User" {
			t.Errorf("expected name 'Test User', got %v", user.Name)
		}
		if user.Age == nil || *user.Age != 25 {
			t.Errorf("expected age 25, got %v", user.Age)
		}
	})

	t.Run("create user with duplicate email returns friendly error", func(t *testing.T) {
		// First user succeeds
		_, err := svc.CreateUser(ctx, "duplicate@example.com",
			sql.NullString{Valid: false},
			sql.NullInt64{Valid: false})
		if err != nil {
			t.Fatalf("first user creation failed: %v", err)
		}

		// Second user with same email fails with friendly message
		_, err = svc.CreateUser(ctx, "duplicate@example.com",
			sql.NullString{Valid: false},
			sql.NullInt64{Valid: false})

		if err == nil {
			t.Fatal("expected error for duplicate email, got nil")
		}
		if !strings.Contains(err.Error(), "duplicate@example.com") {
			t.Errorf("error message should mention email, got: %v", err)
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("error message should say 'already exists', got: %v", err)
		}
	})
}

func TestService_GetUser(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("get existing user", func(t *testing.T) {
		user, err := svc.GetUser(ctx, data.UserID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID != data.UserID1 {
			t.Errorf("expected ID %d, got %d", data.UserID1, user.ID)
		}
	})

	t.Run("get non-existent user returns friendly error", func(t *testing.T) {
		_, err := svc.GetUser(ctx, 99999)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "99999") {
			t.Errorf("error should mention user ID, got: %v", err)
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error should say 'not found', got: %v", err)
		}
	})
}

func TestService_CreateCourse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("create course successfully", func(t *testing.T) {
		course, err := svc.CreateCourse(ctx, "go-101", "Go Programming", 9999)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if course.Slug != "go-101" {
			t.Errorf("expected slug go-101, got %s", course.Slug)
		}
		if course.Price != 9999 {
			t.Errorf("expected price 9999, got %d", course.Price)
		}
	})

	t.Run("create course with duplicate slug returns friendly error", func(t *testing.T) {
		// First course succeeds
		_, err := svc.CreateCourse(ctx, "rust-101", "Rust Programming", 14999)
		if err != nil {
			t.Fatalf("first course creation failed: %v", err)
		}

		// Duplicate fails with friendly message
		_, err = svc.CreateCourse(ctx, "rust-101", "Rust Again", 14999)
		if err == nil {
			t.Fatal("expected error for duplicate slug, got nil")
		}
		if !strings.Contains(err.Error(), "rust-101") {
			t.Errorf("error should mention slug, got: %v", err)
		}
	})
}

func TestService_ListUsers(t *testing.T) {
	db, _ := testutil.SetupTestDBWithData(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := svc.ListUsers(ctx, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}

	// Verify models are converted
	if users[0].Name == nil {
		t.Error("expected name to be set")
	}
}

func TestService_ListCourses(t *testing.T) {
	db, _ := testutil.SetupTestDBWithData(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	courses, err := svc.ListCourses(ctx, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(courses))
	}
}

func TestService_CreateOrder(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("create order with single course", func(t *testing.T) {
		order, err := svc.CreateOrder(ctx, data.UserID1, []int64{data.CourseID1}, "card")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if order.Status != "completed" {
			t.Errorf("expected status completed, got %s", order.Status)
		}
		if order.TotalAmount != 9999 {
			t.Errorf("expected total 9999, got %d", order.TotalAmount)
		}
		if order.PaymentMethod == nil || *order.PaymentMethod != "card" {
			t.Errorf("expected payment method 'card', got %v", order.PaymentMethod)
		}

		// Verify user now owns the course
		owns, err := svc.CheckUserOwnsCourse(ctx, data.UserID1, data.CourseID1)
		if err != nil {
			t.Fatalf("failed to check ownership: %v", err)
		}
		if !owns {
			t.Error("user should own the course after purchase")
		}
	})

	t.Run("create order with multiple courses", func(t *testing.T) {
		order, err := svc.CreateOrder(ctx, data.UserID2,
			[]int64{data.CourseID1, data.CourseID2}, "paypal")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// 9999 + 14999 = 24998
		if order.TotalAmount != 24998 {
			t.Errorf("expected total 24998, got %d", order.TotalAmount)
		}

		// Check order details
		orderWithItems, err := svc.GetOrderWithItems(ctx, order.ID)
		if err != nil {
			t.Fatalf("failed to get order: %v", err)
		}
		if len(orderWithItems.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(orderWithItems.Items))
		}
	})

	t.Run("create order fails with friendly error for non-existent user", func(t *testing.T) {
		_, err := svc.CreateOrder(ctx, 99999, []int64{data.CourseID1}, "card")
		if err == nil {
			t.Fatal("expected error for non-existent user, got nil")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error should mention 'not found', got: %v", err)
		}
	})

	t.Run("create order fails for already enrolled course", func(t *testing.T) {
		// UserID1 already owns CourseID1 from first test
		_, err := svc.CreateOrder(ctx, data.UserID1, []int64{data.CourseID1}, "card")
		if err == nil {
			t.Fatal("expected error for duplicate purchase, got nil")
		}
		if !strings.Contains(err.Error(), "already enrolled") {
			t.Errorf("error should mention 'already enrolled', got: %v", err)
		}
	})
}

func TestService_CheckUserOwnsCourse(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("user does not own course initially", func(t *testing.T) {
		owns, err := svc.CheckUserOwnsCourse(ctx, data.UserID1, data.CourseID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if owns {
			t.Error("user should not own course initially")
		}
	})

	t.Run("user owns course after purchase", func(t *testing.T) {
		// Create order
		_, err := svc.CreateOrder(ctx, data.UserID1, []int64{data.CourseID1}, "card")
		if err != nil {
			t.Fatalf("failed to create order: %v", err)
		}

		// Check ownership
		owns, err := svc.CheckUserOwnsCourse(ctx, data.UserID1, data.CourseID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !owns {
			t.Error("user should own course after purchase")
		}
	})
}

func TestService_GetUserOrders(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("get user orders", func(t *testing.T) {
		// Create two orders
		_, err := svc.CreateOrder(ctx, data.UserID1, []int64{data.CourseID1}, "card")
		if err != nil {
			t.Fatalf("failed to create first order: %v", err)
		}

		_, err = svc.CreateOrder(ctx, data.UserID1, []int64{data.CourseID2}, "paypal")
		if err != nil {
			t.Fatalf("failed to create second order: %v", err)
		}

		// Get orders
		orders, err := svc.GetUserOrders(ctx, data.UserID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(orders))
		}

		// Verify first order has items
		if len(orders[0].Items) == 0 {
			t.Error("expected order to have items")
		}
	})
}

func TestService_EnrollUser(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := repository.New(db)
	svc := New(repo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("enroll user in course (free)", func(t *testing.T) {
		enrollment, err := svc.EnrollUser(ctx, data.UserID1, data.CourseID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if enrollment.UserID != data.UserID1 {
			t.Errorf("expected user ID %d, got %d", data.UserID1, enrollment.UserID)
		}
		if enrollment.Status != "active" {
			t.Errorf("expected status active, got %s", enrollment.Status)
		}
	})

	t.Run("enroll fails with friendly error for duplicate", func(t *testing.T) {
		// UserID1 already enrolled in CourseID1
		_, err := svc.EnrollUser(ctx, data.UserID1, data.CourseID1)
		if err == nil {
			t.Fatal("expected error for duplicate enrollment, got nil")
		}
		if !strings.Contains(err.Error(), "already enrolled") {
			t.Errorf("error should mention 'already enrolled', got: %v", err)
		}
	})
}
