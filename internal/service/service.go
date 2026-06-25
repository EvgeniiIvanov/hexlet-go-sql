package service

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/go-sql/internal/db"
	"example.com/go-sql/internal/models"
	"example.com/go-sql/internal/repository"
)

// Service provides business logic layer
type Service struct {
	repo *repository.Repository
}

// New creates a new service
func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// Course operations

func (s *Service) CreateCourse(ctx context.Context, slug, title string, price int64) (models.Course, error) {
	course, err := s.repo.CreateCourse(ctx, slug, title, price)
	if err != nil {
		if repository.IsConflict(err) {
			return models.Course{}, fmt.Errorf("course with slug '%s' already exists", slug)
		}
		return models.Course{}, fmt.Errorf("failed to create course: %w", err)
	}
	return models.FromDBCourse(course), nil
}

func (s *Service) GetCourse(ctx context.Context, id int64) (models.Course, error) {
	course, err := s.repo.GetCourse(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return models.Course{}, fmt.Errorf("course with id %d not found", id)
		}
		return models.Course{}, fmt.Errorf("failed to get course: %w", err)
	}
	return models.FromDBCourse(course), nil
}

func (s *Service) ListCourses(ctx context.Context, limit, offset int64) ([]models.Course, error) {
	courses, err := s.repo.ListCourses(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list courses: %w", err)
	}
	return models.FromDBCourses(courses), nil
}

func (s *Service) FindCoursesByIDs(ctx context.Context, ids []int64) ([]models.Course, error) {
	courses, err := s.repo.FindCoursesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to find courses: %w", err)
	}
	return models.FromDBCourses(courses), nil
}

func (s *Service) DeleteCourse(ctx context.Context, id int64) error {
	err := s.repo.DeleteCourse(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("course with id %d not found", id)
		}
		return fmt.Errorf("failed to delete course: %w", err)
	}
	return nil
}

// User operations

func (s *Service) CreateUser(ctx context.Context, email string, name sql.NullString, age sql.NullInt64) (models.User, error) {
	user, err := s.repo.CreateUser(ctx, email, name, age)
	if err != nil {
		if repository.IsConflict(err) {
			return models.User{}, fmt.Errorf("user with email '%s' already exists", email)
		}
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return models.FromDBUser(user), nil
}

func (s *Service) GetUser(ctx context.Context, id int64) (models.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return models.User{}, fmt.Errorf("user with id %d not found", id)
		}
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return models.FromDBUser(user), nil
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int64) ([]models.User, error) {
	users, err := s.repo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return models.FromDBUsers(users), nil
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("user with id %d not found", id)
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Enrollment operations

func (s *Service) EnrollUser(ctx context.Context, userID, courseID int64) (models.Enrollment, error) {
	enrollment, err := s.repo.EnrollUser(ctx, userID, courseID)
	if err != nil {
		if repository.IsNotFound(err) {
			return models.Enrollment{}, fmt.Errorf("user or course not found")
		}
		if repository.IsConflict(err) {
			return models.Enrollment{}, fmt.Errorf("user %d is already enrolled in course %d", userID, courseID)
		}
		return models.Enrollment{}, fmt.Errorf("failed to enroll user: %w", err)
	}
	return models.FromDBEnrollment(enrollment), nil
}

func (s *Service) ListEnrollments(ctx context.Context, limit, offset int64) ([]models.EnrollmentWithDetails, error) {
	enrollments, err := s.repo.ListEnrollments(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list enrollments: %w", err)
	}
	return models.FromDBListEnrollmentsRows(enrollments), nil
}

func (s *Service) ListEnrollmentsByUser(ctx context.Context, userID int64) ([]models.EnrollmentWithDetails, error) {
	enrollments, err := s.repo.ListEnrollmentsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user enrollments: %w", err)
	}
	return models.FromDBListEnrollmentsByUserRows(enrollments), nil
}

func (s *Service) ListEnrollmentsByCourse(ctx context.Context, courseID int64) ([]models.EnrollmentWithDetails, error) {
	enrollments, err := s.repo.ListEnrollmentsByCourse(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list course enrollments: %w", err)
	}
	return models.FromDBListEnrollmentsByCourseRows(enrollments), nil
}

func (s *Service) CompleteEnrollment(ctx context.Context, userID, courseID int64) error {
	err := s.repo.CompleteEnrollment(ctx, userID, courseID)
	if err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("active enrollment not found for user %d in course %d", userID, courseID)
		}
		return fmt.Errorf("failed to complete enrollment: %w", err)
	}
	return nil
}

func (s *Service) CancelEnrollment(ctx context.Context, userID, courseID int64) error {
	err := s.repo.CancelEnrollment(ctx, userID, courseID)
	if err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("active enrollment not found for user %d in course %d", userID, courseID)
		}
		return fmt.Errorf("failed to cancel enrollment: %w", err)
	}
	return nil
}

// Bulk operations

func (s *Service) BulkUpsertCourses(ctx context.Context, courses []db.UpsertCourseParams) error {
	return s.repo.BulkUpsertCourses(ctx, courses)
}

func (s *Service) BulkUpsertUsers(ctx context.Context, users []db.UpsertUserParams) error {
	return s.repo.BulkUpsertUsers(ctx, users)
}
