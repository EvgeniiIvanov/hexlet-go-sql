package repository

import "errors"

// Common repository errors
var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrConflict is returned when there's a unique constraint violation
	ErrConflict = errors.New("resource already exists")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// IsNotFound checks if an error is ErrNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsConflict checks if an error is ErrConflict
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsInvalidInput checks if an error is ErrInvalidInput
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}
