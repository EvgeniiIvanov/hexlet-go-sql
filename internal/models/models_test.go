package models

import (
	"database/sql"
	"testing"
	"time"

	"example.com/go-sql/internal/db"
)

func TestFromDBUser(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		dbUser := db.User{
			ID:    1,
			Email: "test@example.com",
			Name: sql.NullString{
				String: "Test User",
				Valid:  true,
			},
			Age: sql.NullInt64{
				Int64: 25,
				Valid: true,
			},
			CreatedAt: sql.NullTime{
				Time:  time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
				Valid: true,
			},
		}

		user := FromDBUser(dbUser)

		if user.ID != 1 {
			t.Errorf("expected ID 1, got %d", user.ID)
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

	t.Run("with null fields", func(t *testing.T) {
		dbUser := db.User{
			ID:    2,
			Email: "minimal@example.com",
			Name:  sql.NullString{Valid: false},
			Age:   sql.NullInt64{Valid: false},
			CreatedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
		}

		user := FromDBUser(dbUser)

		if user.Name != nil {
			t.Errorf("expected nil name, got %v", user.Name)
		}
		if user.Age != nil {
			t.Errorf("expected nil age, got %v", user.Age)
		}
	})
}

func TestFromDBCourse(t *testing.T) {
	dbCourse := db.Course{
		ID:    1,
		Slug:  "go-101",
		Title: "Go Programming",
		Price: 9999,
	}

	course := FromDBCourse(dbCourse)

	if course.ID != 1 {
		t.Errorf("expected ID 1, got %d", course.ID)
	}
	if course.Slug != "go-101" {
		t.Errorf("expected slug go-101, got %s", course.Slug)
	}
	if course.Title != "Go Programming" {
		t.Errorf("expected title 'Go Programming', got %s", course.Title)
	}
	if course.Price != 9999 {
		t.Errorf("expected price 9999, got %d", course.Price)
	}
}

func TestFromDBUsers(t *testing.T) {
	dbUsers := []db.User{
		{ID: 1, Email: "user1@example.com", Name: sql.NullString{String: "User 1", Valid: true}},
		{ID: 2, Email: "user2@example.com", Name: sql.NullString{Valid: false}},
	}

	users := FromDBUsers(dbUsers)

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].ID != 1 {
		t.Errorf("expected first user ID 1, got %d", users[0].ID)
	}
	if users[1].Name != nil {
		t.Errorf("expected second user to have nil name, got %v", users[1].Name)
	}
}

func TestFromDBCourses(t *testing.T) {
	dbCourses := []db.Course{
		{ID: 1, Slug: "go-101", Title: "Go", Price: 9999},
		{ID: 2, Slug: "rust-101", Title: "Rust", Price: 14999},
	}

	courses := FromDBCourses(dbCourses)

	if len(courses) != 2 {
		t.Fatalf("expected 2 courses, got %d", len(courses))
	}

	if courses[0].Price != 9999 {
		t.Errorf("expected first course price 9999, got %d", courses[0].Price)
	}
	if courses[1].Slug != "rust-101" {
		t.Errorf("expected second course slug rust-101, got %s", courses[1].Slug)
	}
}

func TestFromDBOrder(t *testing.T) {
	dbOrder := db.Order{
		ID:          1,
		UserID:      1,
		TotalAmount: 24998,
		Status:      "completed",
		PaymentMethod: sql.NullString{
			String: "card",
			Valid:  true,
		},
		CreatedAt: sql.NullTime{
			Time:  time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC),
			Valid: true,
		},
		CompletedAt: sql.NullTime{
			Time:  time.Date(2026, 6, 26, 10, 1, 0, 0, time.UTC),
			Valid: true,
		},
	}

	order := FromDBOrder(dbOrder)

	if order.ID != 1 {
		t.Errorf("expected ID 1, got %d", order.ID)
	}
	if order.TotalAmount != 24998 {
		t.Errorf("expected total amount 24998, got %d", order.TotalAmount)
	}
	if order.Status != "completed" {
		t.Errorf("expected status completed, got %s", order.Status)
	}
	if order.PaymentMethod == nil || *order.PaymentMethod != "card" {
		t.Errorf("expected payment method 'card', got %v", order.PaymentMethod)
	}
}
