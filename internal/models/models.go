package models

import (
	"encoding/json"
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
	OrderID    *int64    `json:"order_id,omitempty"`
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

	if e.OrderID.Valid {
		model.OrderID = &e.OrderID.Int64
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

// EnrollmentData represents a single enrollment in JSON
type EnrollmentData struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"user_id"`
	UserEmail  string  `json:"user_email"`
	UserName   *string `json:"user_name,omitempty"`
	EnrolledAt string  `json:"enrolled_at"` // Keep as string from SQLite JSON
	Status     string  `json:"status"`
}

// CourseWithEnrollments wraps a course with its enrollments as structured data
type CourseWithEnrollments struct {
	ID          int64            `json:"id"`
	Slug        string           `json:"slug"`
	Title       string           `json:"title"`
	Price       int64            `json:"price"`
	Enrollments []EnrollmentData `json:"enrollments"`
}

// FromDBGetCourseWithEnrollmentsRow converts the sqlc row to a clean model
func FromDBGetCourseWithEnrollmentsRow(row db.GetCourseWithEnrollmentsRow) (CourseWithEnrollments, error) {
	model := CourseWithEnrollments{
		ID:          row.ID,
		Slug:        row.Slug,
		Title:       row.Title,
		Price:       row.Price,
		Enrollments: []EnrollmentData{},
	}

	// Parse the JSON enrollments array
	if row.Enrollments != nil {
		// The interface{} contains JSON string from SQLite
		var jsonStr string
		switch v := row.Enrollments.(type) {
		case string:
			jsonStr = v
		case []byte:
			jsonStr = string(v)
		}

		if jsonStr != "" && jsonStr != "[]" {
			if err := json.Unmarshal([]byte(jsonStr), &model.Enrollments); err != nil {
				return model, err
			}
		}
	}

	return model, nil
}

// Order models

type Order struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	TotalAmount   int64     `json:"total_amount"` // in cents
	Status        string    `json:"status"`
	PaymentMethod *string   `json:"payment_method,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	CompletedAt   *string   `json:"completed_at,omitempty"`
}

func FromDBOrder(o db.Order) Order {
	model := Order{
		ID:          o.ID,
		UserID:      o.UserID,
		TotalAmount: o.TotalAmount,
		Status:      o.Status,
	}

	if o.PaymentMethod.Valid {
		model.PaymentMethod = &o.PaymentMethod.String
	}

	if o.CreatedAt.Valid {
		model.CreatedAt = o.CreatedAt.Time
	}

	if o.CompletedAt.Valid {
		completedAt := o.CompletedAt.Time.Format("2006-01-02 15:04:05")
		model.CompletedAt = &completedAt
	}

	return model
}

func FromDBOrders(orders []db.Order) []Order {
	models := make([]Order, len(orders))
	for i, o := range orders {
		models[i] = FromDBOrder(o)
	}
	return models
}

// OrderItem represents an item in an order
type OrderItemData struct {
	ID          int64  `json:"id"`
	CourseID    int64  `json:"course_id"`
	CourseSlug  string `json:"course_slug"`
	CourseTitle string `json:"course_title"`
	Price       int64  `json:"price"` // in cents
	CreatedAt   string `json:"created_at,omitempty"`
}

// OrderWithItems represents an order with its items
type OrderWithItems struct {
	ID            int64           `json:"id"`
	UserID        int64           `json:"user_id"`
	TotalAmount   int64           `json:"total_amount"` // in cents
	Status        string          `json:"status"`
	PaymentMethod *string         `json:"payment_method,omitempty"`
	CreatedAt     string          `json:"created_at"`
	CompletedAt   *string         `json:"completed_at,omitempty"`
	Items         []OrderItemData `json:"items"`
}

func FromDBGetOrderWithItemsRow(row db.GetOrderWithItemsRow) (OrderWithItems, error) {
	model := OrderWithItems{
		ID:          row.ID,
		UserID:      row.UserID,
		TotalAmount: row.TotalAmount,
		Status:      row.Status,
		Items:       []OrderItemData{},
	}

	if row.PaymentMethod.Valid {
		model.PaymentMethod = &row.PaymentMethod.String
	}

	if row.CreatedAt.Valid {
		createdAt := row.CreatedAt.Time.Format("2006-01-02 15:04:05")
		model.CreatedAt = createdAt
	}

	if row.CompletedAt.Valid {
		completedAt := row.CompletedAt.Time.Format("2006-01-02 15:04:05")
		model.CompletedAt = &completedAt
	}

	// Parse JSON items
	if row.Items != nil {
		var jsonStr string
		switch v := row.Items.(type) {
		case string:
			jsonStr = v
		case []byte:
			jsonStr = string(v)
		}

		if jsonStr != "" && jsonStr != "[]" {
			if err := json.Unmarshal([]byte(jsonStr), &model.Items); err != nil {
				return model, err
			}
		}
	}

	return model, nil
}

func FromDBGetUserOrdersRows(rows []db.GetUserOrdersRow) ([]OrderWithItems, error) {
	models := make([]OrderWithItems, len(rows))
	for i, row := range rows {
		model := OrderWithItems{
			ID:          row.ID,
			UserID:      row.UserID,
			TotalAmount: row.TotalAmount,
			Status:      row.Status,
			Items:       []OrderItemData{},
		}

		if row.PaymentMethod.Valid {
			model.PaymentMethod = &row.PaymentMethod.String
		}

		if row.CreatedAt.Valid {
			createdAt := row.CreatedAt.Time.Format("2006-01-02 15:04:05")
			model.CreatedAt = createdAt
		}

		if row.CompletedAt.Valid {
			completedAt := row.CompletedAt.Time.Format("2006-01-02 15:04:05")
			model.CompletedAt = &completedAt
		}

		// Parse JSON items
		if row.Items != nil {
			var jsonStr string
			switch v := row.Items.(type) {
			case string:
				jsonStr = v
			case []byte:
				jsonStr = string(v)
			}

			if jsonStr != "" && jsonStr != "[]" {
				if err := json.Unmarshal([]byte(jsonStr), &model.Items); err != nil {
					return nil, err
				}
			}
		}

		models[i] = model
	}
	return models, nil
}
