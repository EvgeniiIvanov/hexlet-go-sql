package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"example.com/go-sql/internal/testutil"
)

func TestRepository_CreateUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("create user with all fields", func(t *testing.T) {
		user, err := repo.CreateUser(ctx, "test@example.com",
			sql.NullString{String: "Test User", Valid: true},
			sql.NullInt64{Int64: 25, Valid: true})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", user.Email)
		}
		if !user.Name.Valid || user.Name.String != "Test User" {
			t.Errorf("expected name 'Test User', got %v", user.Name)
		}
	})

	t.Run("create user with duplicate email fails", func(t *testing.T) {
		// First user succeeds
		_, err := repo.CreateUser(ctx, "duplicate@example.com",
			sql.NullString{String: "First", Valid: true},
			sql.NullInt64{Valid: false})
		if err != nil {
			t.Fatalf("first user creation failed: %v", err)
		}

		// Second user with same email should fail with ErrConflict
		_, err = repo.CreateUser(ctx, "duplicate@example.com",
			sql.NullString{String: "Second", Valid: true},
			sql.NullInt64{Valid: false})

		if err == nil {
			t.Fatal("expected error for duplicate email, got nil")
		}
		if !IsConflict(err) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestRepository_GetUser(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("get existing user", func(t *testing.T) {
		user, err := repo.GetUser(ctx, data.UserID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID != data.UserID1 {
			t.Errorf("expected ID %d, got %d", data.UserID1, user.ID)
		}
		if user.Email != "alice@test.com" {
			t.Errorf("expected email alice@test.com, got %s", user.Email)
		}
	})

	t.Run("get non-existent user returns ErrNotFound", func(t *testing.T) {
		_, err := repo.GetUser(ctx, 99999)
		if err == nil {
			t.Fatal("expected error for non-existent user, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRepository_CreateCourse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("create course", func(t *testing.T) {
		course, err := repo.CreateCourse(ctx, "go-101", "Go Programming", 9999)
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

	t.Run("create course with duplicate slug fails", func(t *testing.T) {
		// First course succeeds
		_, err := repo.CreateCourse(ctx, "rust-101", "Rust Programming", 14999)
		if err != nil {
			t.Fatalf("first course creation failed: %v", err)
		}

		// Second course with same slug should fail
		_, err = repo.CreateCourse(ctx, "rust-101", "Rust Again", 14999)
		if err == nil {
			t.Fatal("expected error for duplicate slug, got nil")
		}
		if !IsConflict(err) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestRepository_ListUsers(t *testing.T) {
	db, _ := testutil.SetupTestDBWithData(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("list all users", func(t *testing.T) {
		users, err := repo.ListUsers(ctx, 10, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(users) != 2 {
			t.Errorf("expected 2 users, got %d", len(users))
		}
	})

	t.Run("list with pagination", func(t *testing.T) {
		users, err := repo.ListUsers(ctx, 1, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(users) != 1 {
			t.Errorf("expected 1 user, got %d", len(users))
		}

		users, err = repo.ListUsers(ctx, 1, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(users) != 1 {
			t.Errorf("expected 1 user on second page, got %d", len(users))
		}
	})
}

func TestRepository_CheckUserOwnsCourse(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("user does not own course initially", func(t *testing.T) {
		owns, err := repo.CheckUserOwnsCourse(ctx, data.UserID1, data.CourseID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if owns {
			t.Error("expected user to not own course")
		}
	})
}

func TestRepository_CreateOrderWithEnrollments(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("create order with single course", func(t *testing.T) {
		order, err := repo.CreateOrderWithEnrollments(ctx, data.UserID1, []int64{data.CourseID1}, "card")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if order.Status != "completed" {
			t.Errorf("expected status completed, got %s", order.Status)
		}
		if order.TotalAmount != 9999 {
			t.Errorf("expected total 9999, got %d", order.TotalAmount)
		}

		// Verify enrollment was created
		owns, err := repo.CheckUserOwnsCourse(ctx, data.UserID1, data.CourseID1)
		if err != nil {
			t.Fatalf("failed to check ownership: %v", err)
		}
		if !owns {
			t.Error("user should own the course after purchase")
		}
	})

	t.Run("create order with multiple courses", func(t *testing.T) {
		order, err := repo.CreateOrderWithEnrollments(ctx, data.UserID2,
			[]int64{data.CourseID1, data.CourseID2}, "paypal")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// 9999 + 14999 = 24998
		if order.TotalAmount != 24998 {
			t.Errorf("expected total 24998, got %d", order.TotalAmount)
		}

		// Verify both enrollments were created
		owns1, _ := repo.CheckUserOwnsCourse(ctx, data.UserID2, data.CourseID1)
		owns2, _ := repo.CheckUserOwnsCourse(ctx, data.UserID2, data.CourseID2)
		if !owns1 || !owns2 {
			t.Error("user should own both courses after purchase")
		}
	})

	t.Run("create order fails with non-existent user", func(t *testing.T) {
		_, err := repo.CreateOrderWithEnrollments(ctx, 99999, []int64{data.CourseID1}, "card")
		if err == nil {
			t.Fatal("expected error for non-existent user, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("create order fails with non-existent course", func(t *testing.T) {
		_, err := repo.CreateOrderWithEnrollments(ctx, data.UserID1, []int64{99999}, "card")
		if err == nil {
			t.Fatal("expected error for non-existent course, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("create order fails for already purchased course", func(t *testing.T) {
		// UserID1 already owns CourseID1 from first test
		_, err := repo.CreateOrderWithEnrollments(ctx, data.UserID1, []int64{data.CourseID1}, "card")
		if err == nil {
			t.Fatal("expected error for duplicate purchase, got nil")
		}
		if !IsConflict(err) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestRepository_EnrollUser(t *testing.T) {
	db, data := testutil.SetupTestDBWithData(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("enroll user in course (free enrollment)", func(t *testing.T) {
		enrollment, err := repo.EnrollUser(ctx, data.UserID1, data.CourseID1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if enrollment.UserID != data.UserID1 {
			t.Errorf("expected user ID %d, got %d", data.UserID1, enrollment.UserID)
		}
		if enrollment.CourseID != data.CourseID1 {
			t.Errorf("expected course ID %d, got %d", data.CourseID1, enrollment.CourseID)
		}
		if enrollment.Status != "active" {
			t.Errorf("expected status active, got %s", enrollment.Status)
		}
		if enrollment.OrderID.Valid {
			t.Error("expected order ID to be NULL for free enrollment")
		}
	})

	t.Run("enroll fails with non-existent user", func(t *testing.T) {
		_, err := repo.EnrollUser(ctx, 99999, data.CourseID1)
		if err == nil {
			t.Fatal("expected error for non-existent user, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("enroll fails with non-existent course", func(t *testing.T) {
		_, err := repo.EnrollUser(ctx, data.UserID2, 99999)
		if err == nil {
			t.Fatal("expected error for non-existent course, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("enroll fails for duplicate enrollment", func(t *testing.T) {
		// UserID1 already enrolled in CourseID1
		_, err := repo.EnrollUser(ctx, data.UserID1, data.CourseID1)
		if err == nil {
			t.Fatal("expected error for duplicate enrollment, got nil")
		}
		if !IsConflict(err) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}
