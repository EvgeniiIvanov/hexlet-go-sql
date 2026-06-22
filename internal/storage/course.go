package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// CreateCourseDTO represents data for creating a new course
type CreateCourseDTO struct {
	Slug  string
	Title string
	Price int
}

// UpdateCourseDTO represents data for updating a course
type UpdateCourseDTO struct {
	ID    int64
	Slug  *string // Nullable - if nil, field won't be updated
	Title *string // Nullable - if nil, field won't be updated
	Price *int    // Nullable - if nil, field won't be updated
}

// CourseRepository handles all database operations for courses
type CourseRepository struct {
	db *sql.DB
}

// NewCourseRepository creates a new course repository
func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

// Create inserts a new course into the database
func (r *CourseRepository) Create(ctx context.Context, dto CreateCourseDTO) (Course, error) {
	const query = `
		INSERT INTO courses (slug, title, price)
		VALUES (?, ?, ?)
		RETURNING id, slug, title, price
	`

	var course Course
	err := r.db.QueryRowContext(ctx, query, dto.Slug, dto.Title, dto.Price).
		Scan(&course.ID, &course.Slug, &course.Title, &course.Price)
	if err != nil {
		return Course{}, fmt.Errorf("create course: %w", err)
	}

	return course, nil
}

// List returns a paginated list of courses
func (r *CourseRepository) List(ctx context.Context, limit, offset int, orderBy string) ([]Course, error) {
	// Allowed order clauses to prevent SQL injection
	allowedOrder := map[string]string{
		"id":         "id ASC",
		"slug":       "slug ASC",
		"title":      "title ASC",
		"price":      "price ASC",
		"price_asc":  "price ASC",
		"price_desc": "price DESC",
		"title_asc":  "title ASC",
	}

	order, ok := allowedOrder[orderBy]
	if !ok {
		order = "id ASC"
	}

	query := `
		SELECT id, slug, title, price
		FROM courses
		ORDER BY ` + order + `
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, c)
	}

	return courses, rows.Err()
}

// FindByIDs returns courses matching the given IDs
func (r *CourseRepository) FindByIDs(ctx context.Context, ids []int64) ([]Course, error) {
	if len(ids) == 0 {
		return []Course{}, nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	query := `
		SELECT id, slug, title, price
		FROM courses
		WHERE id IN (` + placeholders + `)
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find courses: %w", err)
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, c)
	}

	return courses, rows.Err()
}

// Get retrieves a course by ID
func (r *CourseRepository) Get(ctx context.Context, id int64) (Course, error) {
	const query = `SELECT id, slug, title, price FROM courses WHERE id = ?`

	var course Course
	err := r.db.QueryRowContext(ctx, query, id).Scan(&course.ID, &course.Slug, &course.Title, &course.Price)
	if err == sql.ErrNoRows {
		return Course{}, nil // Not found
	}
	if err != nil {
		return Course{}, fmt.Errorf("get course: %w", err)
	}

	return course, nil
}

// FindBySlug returns a course by its slug
func (r *CourseRepository) FindBySlug(ctx context.Context, slug string) (*Course, error) {
	const query = `SELECT id, slug, title, price FROM courses WHERE slug = ?`

	var course Course
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&course.ID, &course.Slug, &course.Title, &course.Price)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("find course by slug: %w", err)
	}

	return &course, nil
}

// Update modifies an existing course
// Only updates fields that are non-nil in the DTO using COALESCE
func (r *CourseRepository) Update(ctx context.Context, dto UpdateCourseDTO) (Course, error) {
	const query = `
		UPDATE courses
		SET slug = COALESCE(?, slug),
			title = COALESCE(?, title),
			price = COALESCE(?, price)
		WHERE id = ?
		RETURNING id, slug, title, price
	`

	var course Course
	err := r.db.QueryRowContext(ctx, query, dto.Slug, dto.Title, dto.Price, dto.ID).
		Scan(&course.ID, &course.Slug, &course.Title, &course.Price)
	if err != nil {
		return Course{}, fmt.Errorf("update course: %w", err)
	}

	return course, nil
}

// Delete removes a course by ID
func (r *CourseRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM courses WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete course: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("course not found")
	}

	return nil
}

// BulkUpsert inserts or updates multiple courses using prepared statements
// If a course with the same slug exists, updates title and price
func (r *CourseRepository) BulkUpsert(ctx context.Context, dtos []CreateCourseDTO) error {
	if len(dtos) == 0 {
		return nil
	}

	stmt, err := r.db.PrepareContext(ctx, `
		INSERT INTO courses(slug, title, price)
		VALUES(?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			title = excluded.title,
			price = excluded.price
	`)
	if err != nil {
		return fmt.Errorf("prepare bulk upsert: %w", err)
	}
	defer stmt.Close()

	for _, dto := range dtos {
		if _, err := stmt.ExecContext(ctx, dto.Slug, dto.Title, dto.Price); err != nil {
			return fmt.Errorf("upsert course %s: %w", dto.Slug, err)
		}
	}

	return nil
}

// BulkCreate inserts multiple courses using prepared statements
func (r *CourseRepository) BulkCreate(ctx context.Context, dtos []CreateCourseDTO) ([]Course, error) {
	if len(dtos) == 0 {
		return []Course{}, nil
	}

	stmt, err := r.db.PrepareContext(ctx, `
		INSERT INTO courses(slug, title, price)
		VALUES(?, ?, ?)
		RETURNING id, slug, title, price
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare bulk create: %w", err)
	}
	defer stmt.Close()

	courses := make([]Course, 0, len(dtos))
	for _, dto := range dtos {
		var course Course
		err := stmt.QueryRowContext(ctx, dto.Slug, dto.Title, dto.Price).
			Scan(&course.ID, &course.Slug, &course.Title, &course.Price)
		if err != nil {
			return nil, fmt.Errorf("create course %s: %w", dto.Slug, err)
		}
		courses = append(courses, course)
	}

	return courses, nil
}
