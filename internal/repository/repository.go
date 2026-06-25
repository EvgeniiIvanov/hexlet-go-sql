package repository

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/go-sql/internal/db"
)

// Repository wraps sqlc queries and provides transaction support
type Repository struct {
	db      *sql.DB
	queries *db.Queries
}

// New creates a new repository
func New(database *sql.DB) *Repository {
	return &Repository{
		db:      database,
		queries: db.New(database),
	}
}

// Queries returns the underlying sqlc queries for direct access
func (r *Repository) Queries() *db.Querier {
	var q db.Querier = r.queries
	return &q
}

// withTx executes a function within a database transaction
func (r *Repository) withTx(ctx context.Context, fn func(*db.Queries) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit()
}

// EnrollUser enrolls a user in a course using a transaction
// This ensures atomicity: user/course existence checks + enrollment creation
func (r *Repository) EnrollUser(ctx context.Context, userID, courseID int64) (db.Enrollment, error) {
	var enrollment db.Enrollment

	err := r.withTx(ctx, func(q *db.Queries) error {
		// Check user exists
		_, err := q.CheckUserExists(ctx, userID)
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %d not found", userID)
		}
		if err != nil {
			return fmt.Errorf("check user exists: %w", err)
		}

		// Check course exists
		_, err = q.CheckCourseExists(ctx, courseID)
		if err == sql.ErrNoRows {
			return fmt.Errorf("course with id %d not found", courseID)
		}
		if err != nil {
			return fmt.Errorf("check course exists: %w", err)
		}

		// Create enrollment
		enrollment, err = q.CreateEnrollment(ctx, db.CreateEnrollmentParams{
			UserID:   userID,
			CourseID: courseID,
			Status:   "active",
		})
		if err != nil {
			return fmt.Errorf("create enrollment: %w", err)
		}

		return nil
	})

	return enrollment, err
}

// BulkUpsertUsers performs bulk upsert of users using prepared statements
func (r *Repository) BulkUpsertUsers(ctx context.Context, users []db.UpsertUserParams) error {
	return r.withTx(ctx, func(q *db.Queries) error {
		for _, u := range users {
			if err := q.UpsertUser(ctx, u); err != nil {
				return fmt.Errorf("upsert user %s: %w", u.Email, err)
			}
		}
		return nil
	})
}

// BulkUpsertCourses performs bulk upsert of courses using prepared statements
func (r *Repository) BulkUpsertCourses(ctx context.Context, courses []db.UpsertCourseParams) error {
	return r.withTx(ctx, func(q *db.Queries) error {
		for _, c := range courses {
			if err := q.UpsertCourse(ctx, c); err != nil {
				return fmt.Errorf("upsert course %s: %w", c.Slug, err)
			}
		}
		return nil
	})
}

// CompleteEnrollment marks an enrollment as completed
func (r *Repository) CompleteEnrollment(ctx context.Context, userID, courseID int64) error {
	err := r.queries.UpdateEnrollmentStatus(ctx, db.UpdateEnrollmentStatusParams{
		Status:   "completed",
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return fmt.Errorf("complete enrollment: %w", err)
	}
	return nil
}

// CancelEnrollment marks an enrollment as cancelled
func (r *Repository) CancelEnrollment(ctx context.Context, userID, courseID int64) error {
	err := r.queries.UpdateEnrollmentStatus(ctx, db.UpdateEnrollmentStatusParams{
		Status:   "cancelled",
		UserID:   userID,
		CourseID: courseID,
	})
	if err != nil {
		return fmt.Errorf("cancel enrollment: %w", err)
	}
	return nil
}
