package storage

type Course struct {
	ID    int64  `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

type User struct {
	ID        int64   `json:"id"`
	Email     string  `json:"email"`
	Name      *string `json:"name,omitempty"` // Nullable
	Age       *int    `json:"age,omitempty"`  // Nullable
	CreatedAt string  `json:"created_at"`
}

type Enrollment struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	CourseID   int64  `json:"course_id"`
	EnrolledAt string `json:"enrolled_at"`
	Status     string `json:"status"` // active, completed, cancelled
}

// EnrollmentWithDetails includes user and course information
type EnrollmentWithDetails struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	UserEmail   string  `json:"user_email"`
	UserName    *string `json:"user_name,omitempty"`
	CourseID    int64   `json:"course_id"`
	CourseSlug  string  `json:"course_slug"`
	CourseTitle string  `json:"course_title"`
	EnrolledAt  string  `json:"enrolled_at"`
	Status      string  `json:"status"`
}
