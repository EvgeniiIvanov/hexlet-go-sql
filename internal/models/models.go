package models

import (
	"time"

	"example.com/go-sql/internal/db"
)

// User wraps the sqlc User type with better JSON marshaling
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name,omitempty"`
	Age       *int64    `json:"age,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FromDBUser converts sqlc User to our User model
func FromDBUser(u db.User) User {
	model := User{
		ID:    u.ID,
		Email: u.Email,
	}

	if u.Name.Valid {
		model.Name = &u.Name.String
	}

	if u.Age.Valid {
		model.Age = &u.Age.Int64
	}

	if u.CreatedAt.Valid {
		model.CreatedAt = u.CreatedAt.Time
	}

	return model
}

// FromDBUsers converts a slice of sqlc Users to our User models
func FromDBUsers(users []db.User) []User {
	models := make([]User, len(users))
	for i, u := range users {
		models[i] = FromDBUser(u)
	}
	return models
}

// Course wraps the sqlc Course type
type Course struct {
	ID    int64  `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int64  `json:"price"`
}

// FromDBCourse converts sqlc Course to our Course model
func FromDBCourse(c db.Course) Course {
	return Course{
		ID:    c.ID,
		Slug:  c.Slug,
		Title: c.Title,
		Price: c.Price,
	}
}

// FromDBCourses converts a slice of sqlc Courses to our Course models
func FromDBCourses(courses []db.Course) []Course {
	models := make([]Course, len(courses))
	for i, c := range courses {
		models[i] = FromDBCourse(c)
	}
	return models
}

// Enrollment wraps the sqlc Enrollment type with better JSON marshaling
type Enrollment struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	CourseID   int64     `json:"course_id"`
	EnrolledAt time.Time `json:"enrolled_at"`
	Status     string    `json:"status"`
}

// FromDBEnrollment converts sqlc Enrollment to our Enrollment model
func FromDBEnrollment(e db.Enrollment) Enrollment {
	model := Enrollment{
		ID:       e.ID,
		UserID:   e.UserID,
		CourseID: e.CourseID,
		Status:   e.Status,
	}

	if e.EnrolledAt.Valid {
		model.EnrolledAt = e.EnrolledAt.Time
	}

	return model
}

// EnrollmentWithDetails contains enrollment info with user and course details
type EnrollmentWithDetails struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	UserName    *string   `json:"user_name,omitempty"`
	CourseID    int64     `json:"course_id"`
	CourseSlug  string    `json:"course_slug"`
	CourseTitle string    `json:"course_title"`
	EnrolledAt  time.Time `json:"enrolled_at"`
	Status      string    `json:"status"`
}

// FromDBListEnrollmentsRow converts sqlc ListEnrollmentsRow to our model
func FromDBListEnrollmentsRow(row db.ListEnrollmentsRow) EnrollmentWithDetails {
	model := EnrollmentWithDetails{
		ID:          row.ID,
		UserID:      row.UserID,
		UserEmail:   row.UserEmail,
		CourseID:    row.CourseID,
		CourseSlug:  row.CourseSlug,
		CourseTitle: row.CourseTitle,
		Status:      row.Status,
	}

	if row.UserName.Valid {
		model.UserName = &row.UserName.String
	}

	if row.EnrolledAt.Valid {
		model.EnrolledAt = row.EnrolledAt.Time
	}

	return model
}

// FromDBListEnrollmentsRows converts a slice
func FromDBListEnrollmentsRows(rows []db.ListEnrollmentsRow) []EnrollmentWithDetails {
	models := make([]EnrollmentWithDetails, len(rows))
	for i, row := range rows {
		models[i] = FromDBListEnrollmentsRow(row)
	}
	return models
}

// FromDBListEnrollmentsByUserRow converts sqlc ListEnrollmentsByUserRow to our model
func FromDBListEnrollmentsByUserRow(row db.ListEnrollmentsByUserRow) EnrollmentWithDetails {
	model := EnrollmentWithDetails{
		ID:          row.ID,
		UserID:      row.UserID,
		UserEmail:   row.UserEmail,
		CourseID:    row.CourseID,
		CourseSlug:  row.CourseSlug,
		CourseTitle: row.CourseTitle,
		Status:      row.Status,
	}

	if row.UserName.Valid {
		model.UserName = &row.UserName.String
	}

	if row.EnrolledAt.Valid {
		model.EnrolledAt = row.EnrolledAt.Time
	}

	return model
}

// FromDBListEnrollmentsByUserRows converts a slice
func FromDBListEnrollmentsByUserRows(rows []db.ListEnrollmentsByUserRow) []EnrollmentWithDetails {
	models := make([]EnrollmentWithDetails, len(rows))
	for i, row := range rows {
		models[i] = FromDBListEnrollmentsByUserRow(row)
	}
	return models
}

// FromDBListEnrollmentsByCourseRow converts sqlc ListEnrollmentsByCourseRow to our model
func FromDBListEnrollmentsByCourseRow(row db.ListEnrollmentsByCourseRow) EnrollmentWithDetails {
	model := EnrollmentWithDetails{
		ID:          row.ID,
		UserID:      row.UserID,
		UserEmail:   row.UserEmail,
		CourseID:    row.CourseID,
		CourseSlug:  row.CourseSlug,
		CourseTitle: row.CourseTitle,
		Status:      row.Status,
	}

	if row.UserName.Valid {
		model.UserName = &row.UserName.String
	}

	if row.EnrolledAt.Valid {
		model.EnrolledAt = row.EnrolledAt.Time
	}

	return model
}

// FromDBListEnrollmentsByCourseRows converts a slice
func FromDBListEnrollmentsByCourseRows(rows []db.ListEnrollmentsByCourseRow) []EnrollmentWithDetails {
	models := make([]EnrollmentWithDetails, len(rows))
	for i, row := range rows {
		models[i] = FromDBListEnrollmentsByCourseRow(row)
	}
	return models
}
