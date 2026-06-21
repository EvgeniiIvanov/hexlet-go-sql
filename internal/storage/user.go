package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// CreateUserDTO represents data for creating a new user
type CreateUserDTO struct {
	Email string
	Name  *string // Nullable
	Age   *int    // Nullable
}

// UpdateUserDTO represents data for updating a user
type UpdateUserDTO struct {
	ID   int64
	Name *string // Nullable - if nil, field won't be updated
	Age  *int    // Nullable - if nil, field won't be updated
}

// UserRepository handles all database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, dto CreateUserDTO) (User, error) {
	const query = `
		INSERT INTO users(email, name, age)
		VALUES(?, ?, ?)
		RETURNING id, email, name, age, created_at
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, dto.Email, dto.Name, dto.Age).
		Scan(&user.ID, &user.Email, &user.Name, &user.Age, &user.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// Update modifies an existing user
// Only updates fields that are non-nil in the DTO using COALESCE
func (r *UserRepository) Update(ctx context.Context, dto UpdateUserDTO) (User, error) {
	const query = `
		UPDATE users
		SET name = COALESCE(?, name),
			age  = COALESCE(?, age)
		WHERE id = ?
		RETURNING id, email, name, age, created_at
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, dto.Name, dto.Age, dto.ID).
		Scan(&user.ID, &user.Email, &user.Name, &user.Age, &user.CreatedAt)
	if err != nil {
		return User{}, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

// Get retrieves a user by ID
func (r *UserRepository) Get(ctx context.Context, id int64) (User, error) {
	const query = `
		SELECT id, email, name, age, created_at
		FROM users
		WHERE id = ?
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.Name, &user.Age, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return User{}, nil // Not found
	}
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

// List returns all users ordered by ID
func (r *UserRepository) List(ctx context.Context) ([]User, error) {
	const query = `
		SELECT id, email, name, age, created_at
		FROM users
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Age, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// FindByEmail retrieves a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, email, name, age, created_at
		FROM users
		WHERE email = ?
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Name, &user.Age, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}
