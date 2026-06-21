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
