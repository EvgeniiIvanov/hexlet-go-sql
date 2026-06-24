package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// EnrollmentRepository handles enrollment-related database operations
type EnrollmentRepository struct {
	db *sql.DB
}

// NewEnrollmentRepository creates a new enrollment repository
func NewEnrollmentRepository(db *sql.DB) *EnrollmentRepository {
	return &EnrollmentRepository{db: db}
}

// withTx executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// EnrollUser enrolls a user in a course using a transaction
// This ensures atomicity: either the enrollment is created successfully,
// or nothing happens (in case of errors like user/course not found)
func (r *EnrollmentRepository) EnrollUser(ctx context.Context, userID, courseID int64) (*Enrollment, error) {
	var enrollment Enrollment

	err := withTx(ctx, r.db, func(tx *sql.Tx) error {
		// 1. Verify user exists
		var userExists bool
		err := tx.QueryRowContext(ctx, "SELECT 1 FROM users WHERE id = ?", userID).Scan(&userExists)
		if err == sql.ErrNoRows {
			return fmt.Errorf("user with id %d not found", userID)
		}
		if err != nil {
			return fmt.Errorf("check user exists: %w", err)
		}

		// 2. Verify course exists
		var courseExists bool
		err = tx.QueryRowContext(ctx, "SELECT 1 FROM courses WHERE id = ?", courseID).Scan(&courseExists)
		if err == sql.ErrNoRows {
			return fmt.Errorf("course with id %d not found", courseID)
		}
		if err != nil {
			return fmt.Errorf("check course exists: %w", err)
		}

		// 3. Create enrollment (UNIQUE constraint prevents duplicate enrollments)
		err = tx.QueryRowContext(ctx, `
			INSERT INTO enrollments(user_id, course_id, status)
			VALUES(?, ?, 'active')
			RETURNING id, user_id, course_id, enrolled_at, status
		`, userID, courseID).Scan(
			&enrollment.ID,
			&enrollment.UserID,
			&enrollment.CourseID,
			&enrollment.EnrolledAt,
			&enrollment.Status,
		)
		if err != nil {
			return fmt.Errorf("create enrollment: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &enrollment, nil
}

// UnenrollUser removes a user from a course (sets status to 'cancelled')
func (r *EnrollmentRepository) UnenrollUser(ctx context.Context, userID, courseID int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE enrollments 
		SET status = 'cancelled'
		WHERE user_id = ? AND course_id = ? AND status = 'active'
	`, userID, courseID)
	if err != nil {
		return fmt.Errorf("unenroll user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no active enrollment found for user %d in course %d", userID, courseID)
	}

	return nil
}

// GetUserEnrollments returns all enrollments for a specific user with course details
func (r *EnrollmentRepository) GetUserEnrollments(ctx context.Context, userID int64) ([]EnrollmentWithDetails, error) {
	query := `
		SELECT 
			e.id, e.user_id, u.email, u.name,
			e.course_id, c.slug, c.title,
			e.enrolled_at, e.status
		FROM enrollments e
		JOIN users u ON e.user_id = u.id
		JOIN courses c ON e.course_id = c.id
		WHERE e.user_id = ?
		ORDER BY e.enrolled_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []EnrollmentWithDetails
	for rows.Next() {
		var e EnrollmentWithDetails
		err := rows.Scan(
			&e.ID, &e.UserID, &e.UserEmail, &e.UserName,
			&e.CourseID, &e.CourseSlug, &e.CourseTitle,
			&e.EnrolledAt, &e.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, rows.Err()
}

// GetCourseEnrollments returns all enrollments for a specific course with user details
func (r *EnrollmentRepository) GetCourseEnrollments(ctx context.Context, courseID int64) ([]EnrollmentWithDetails, error) {
	query := `
		SELECT
			e.id, e.user_id, u.email, u.name,
			e.course_id, c.slug, c.title,
			e.enrolled_at, e.status
		FROM enrollments e
		JOIN users u ON e.user_id = u.id
		JOIN courses c ON e.course_id = c.id
		WHERE e.course_id = ?
		ORDER BY e.enrolled_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("query enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []EnrollmentWithDetails
	for rows.Next() {
		var e EnrollmentWithDetails
		err := rows.Scan(
			&e.ID, &e.UserID, &e.UserEmail, &e.UserName,
			&e.CourseID, &e.CourseSlug, &e.CourseTitle,
			&e.EnrolledAt, &e.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, rows.Err()
}

// ListAll returns all enrollments with details
func (r *EnrollmentRepository) ListAll(ctx context.Context) ([]EnrollmentWithDetails, error) {
	query := `
		SELECT
			e.id, e.user_id, u.email, u.name,
			e.course_id, c.slug, c.title,
			e.enrolled_at, e.status
		FROM enrollments e
		JOIN users u ON e.user_id = u.id
		JOIN courses c ON e.course_id = c.id
		ORDER BY e.enrolled_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []EnrollmentWithDetails
	for rows.Next() {
		var e EnrollmentWithDetails
		err := rows.Scan(
			&e.ID, &e.UserID, &e.UserEmail, &e.UserName,
			&e.CourseID, &e.CourseSlug, &e.CourseTitle,
			&e.EnrolledAt, &e.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, rows.Err()
}

// CompleteEnrollment marks an enrollment as completed
func (r *EnrollmentRepository) CompleteEnrollment(ctx context.Context, userID, courseID int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE enrollments
		SET status = 'completed'
		WHERE user_id = ? AND course_id = ? AND status = 'active'
	`, userID, courseID)
	if err != nil {
		return fmt.Errorf("complete enrollment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no active enrollment found for user %d in course %d", userID, courseID)
	}

	return nil
}
