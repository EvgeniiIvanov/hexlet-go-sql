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

// NewWithQueries creates a new repository with provided queries
// This is useful for integration tests where you want to use a transaction
func NewWithQueries(queries *db.Queries) *Repository {
	return &Repository{
		db:      nil, // db is nil when using existing queries/transaction
		queries: queries,
	}
}

// withTx executes a function within a database transaction
func (r *Repository) withTx(ctx context.Context, fn func(*db.Queries) error) error {
	// If db is nil, we're already in a transaction (using NewWithQueries)
	// In this case, just execute the function with the existing queries
	if r.db == nil {
		return fn(r.queries)
	}

	// Otherwise, create a new transaction
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

func (r *Repository) GetCourseWithEnrollments(ctx context.Context, id int64) (db.GetCourseWithEnrollmentsRow, error) {
	course, err := r.queries.GetCourseWithEnrollments(ctx, db.GetCourseWithEnrollmentsParams{
		CourseID: id,
		ID:       id,
	})
	return course, mapError(err)
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

		// Create enrollment (free enrollment, no order)
		enrollment, err = q.CreateEnrollment(ctx, db.CreateEnrollmentParams{
			UserID:   userID,
			CourseID: courseID,
			Status:   "active",
			OrderID:  sql.NullInt64{Valid: false}, // No order for free enrollments
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

// Order operations

func (r *Repository) GetOrder(ctx context.Context, id int64) (db.Order, error) {
	order, err := r.queries.GetOrder(ctx, id)
	return order, mapError(err)
}

func (r *Repository) ListOrders(ctx context.Context, limit, offset int64) ([]db.Order, error) {
	orders, err := r.queries.ListOrders(ctx, db.ListOrdersParams{
		Limit:  limit,
		Offset: offset,
	})
	return orders, mapError(err)
}

func (r *Repository) GetUserOrders(ctx context.Context, userID int64) ([]db.GetUserOrdersRow, error) {
	orders, err := r.queries.GetUserOrders(ctx, userID)
	return orders, mapError(err)
}

func (r *Repository) GetOrderWithItems(ctx context.Context, id int64) (db.GetOrderWithItemsRow, error) {
	order, err := r.queries.GetOrderWithItems(ctx, db.GetOrderWithItemsParams{
		OrderID: id,
		ID:      id,
	})
	return order, mapError(err)
}

func (r *Repository) CheckUserOwnsCourse(ctx context.Context, userID, courseID int64) (bool, error) {
	result, err := r.queries.CheckUserOwnsCourse(ctx, db.CheckUserOwnsCourseParams{
		UserID:   userID,
		CourseID: courseID,
	})
	return result, mapError(err)
}

// CreateOrderWithEnrollments creates an order, order items, and enrollments in a transaction
// This is the key transactional operation for purchase flow
func (r *Repository) CreateOrderWithEnrollments(ctx context.Context, userID int64, courseIDs []int64, paymentMethod string) (db.Order, error) {
	var order db.Order

	err := r.withTx(ctx, func(q *db.Queries) error {
		// 1. Check user exists
		_, err := q.CheckUserExists(ctx, userID)
		if err != nil {
			return mapError(err)
		}

		// 2. Get course prices and validate all courses exist
		var totalAmount int64
		coursePrices := make(map[int64]int64)
		for _, courseID := range courseIDs {
			course, err := q.GetCourse(ctx, courseID)
			if err != nil {
				return mapError(err)
			}
			coursePrices[courseID] = course.Price
			totalAmount += course.Price
		}

		// 3. Create order
		order, err = q.CreateOrder(ctx, db.CreateOrderParams{
			UserID:      userID,
			TotalAmount: totalAmount,
			Status:      "pending",
			PaymentMethod: sql.NullString{
				String: paymentMethod,
				Valid:  paymentMethod != "",
			},
		})
		if err != nil {
			return mapError(err)
		}

		// 4. Create order items
		for _, courseID := range courseIDs {
			_, err := q.CreateOrderItem(ctx, db.CreateOrderItemParams{
				OrderID:  order.ID,
				CourseID: courseID,
				Price:    coursePrices[courseID],
			})
			if err != nil {
				return mapError(err)
			}
		}

		// 5. Mark order as completed (simulating successful payment)
		err = q.CompleteOrder(ctx, order.ID)
		if err != nil {
			return mapError(err)
		}

		// 6. Create enrollments for each course
		for _, courseID := range courseIDs {
			_, err := q.CreateEnrollment(ctx, db.CreateEnrollmentParams{
				UserID:   userID,
				CourseID: courseID,
				Status:   "active",
				OrderID: sql.NullInt64{
					Int64: order.ID,
					Valid: true,
				},
			})
			if err != nil {
				return mapError(err)
			}
		}

		// Update order with completed status
		order.Status = "completed"

		return nil
	})

	return order, err
}

// RefundOrderWithEnrollments refunds an order and cancels associated enrollments
func (r *Repository) RefundOrderWithEnrollments(ctx context.Context, orderID int64) error {
	return r.withTx(ctx, func(q *db.Queries) error {
		// 1. Get order to verify it exists and is completed
		order, err := q.GetOrder(ctx, orderID)
		if err != nil {
			return mapError(err)
		}

		if order.Status != "completed" {
			return fmt.Errorf("can only refund completed orders, current status: %s", order.Status)
		}

		// 2. Mark order as refunded
		err = q.RefundOrder(ctx, orderID)
		if err != nil {
			return mapError(err)
		}

		// 3. Cancel all enrollments associated with this order
		// Note: We'd need to add a query for this, but for now we'll document it
		// TODO: Add CancelEnrollmentsByOrder query

		return nil
	})
}
