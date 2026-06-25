package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

// mapError converts database errors to repository errors
func mapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "UNIQUE constraint failed") || strings.Contains(errMsg, "constraint failed") {
		return ErrConflict
	}

	return err
}

// Course operations

func (r *Repository) CreateCourse(ctx context.Context, slug, title string, price int64) (db.Course, error) {
	course, err := r.queries.CreateCourse(ctx, db.CreateCourseParams{
		Slug:  slug,
		Title: title,
		Price: price,
	})
	return course, mapError(err)
}

func (r *Repository) GetCourse(ctx context.Context, id int64) (db.Course, error) {
	course, err := r.queries.GetCourse(ctx, id)
	return course, mapError(err)
}

func (r *Repository) ListCourses(ctx context.Context, limit, offset int64) ([]db.Course, error) {
	courses, err := r.queries.ListCourses(ctx, db.ListCoursesParams{
		Limit:  limit,
		Offset: offset,
	})
	return courses, mapError(err)
}

func (r *Repository) FindCoursesByIDs(ctx context.Context, ids []int64) ([]db.Course, error) {
	courses, err := r.queries.FindCoursesByIDs(ctx, ids)
	return courses, mapError(err)
}

func (r *Repository) DeleteCourse(ctx context.Context, id int64) error {
	err := r.queries.DeleteCourse(ctx, id)
	return mapError(err)
}

// User operations

func (r *Repository) CreateUser(ctx context.Context, email string, name sql.NullString, age sql.NullInt64) (db.User, error) {
	user, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email: email,
		Name:  name,
		Age:   age,
	})
	return user, mapError(err)
}

func (r *Repository) GetUser(ctx context.Context, id int64) (db.User, error) {
	user, err := r.queries.GetUser(ctx, id)
	return user, mapError(err)
}

func (r *Repository) ListUsers(ctx context.Context, limit, offset int64) ([]db.User, error) {
	users, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	return users, mapError(err)
}

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	err := r.queries.DeleteUser(ctx, id)
	return mapError(err)
}

// Enrollment operations

func (r *Repository) EnrollUser(ctx context.Context, userID, courseID int64) (db.Enrollment, error) {
	var enrollment db.Enrollment

	err := r.withTx(ctx, func(q *db.Queries) error {
		// Check user exists
		_, err := q.CheckUserExists(ctx, userID)
		if err != nil {
			return mapError(err)
		}

		// Check course exists
		_, err = q.CheckCourseExists(ctx, courseID)
		if err != nil {
			return mapError(err)
		}

		// Create enrollment
		enrollment, err = q.CreateEnrollment(ctx, db.CreateEnrollmentParams{
			UserID:   userID,
			CourseID: courseID,
			Status:   "active",
		})
		return mapError(err)
	})

	return enrollment, err
}

func (r *Repository) ListEnrollments(ctx context.Context, limit, offset int64) ([]db.ListEnrollmentsRow, error) {
	enrollments, err := r.queries.ListEnrollments(ctx, db.ListEnrollmentsParams{
		Limit:  limit,
		Offset: offset,
	})
	return enrollments, mapError(err)
}

func (r *Repository) ListEnrollmentsByUser(ctx context.Context, userID int64) ([]db.ListEnrollmentsByUserRow, error) {
	enrollments, err := r.queries.ListEnrollmentsByUser(ctx, userID)
	return enrollments, mapError(err)
}

func (r *Repository) ListEnrollmentsByCourse(ctx context.Context, courseID int64) ([]db.ListEnrollmentsByCourseRow, error) {
	enrollments, err := r.queries.ListEnrollmentsByCourse(ctx, courseID)
	return enrollments, mapError(err)
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

func (r *Repository) CompleteEnrollment(ctx context.Context, userID, courseID int64) error {
	err := r.queries.UpdateEnrollmentStatus(ctx, db.UpdateEnrollmentStatusParams{
		Status:   "completed",
		UserID:   userID,
		CourseID: courseID,
	})
	return mapError(err)
}

func (r *Repository) CancelEnrollment(ctx context.Context, userID, courseID int64) error {
	err := r.queries.UpdateEnrollmentStatus(ctx, db.UpdateEnrollmentStatusParams{
		Status:   "cancelled",
		UserID:   userID,
		CourseID: courseID,
	})
	return mapError(err)
}
