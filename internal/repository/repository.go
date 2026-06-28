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
	db         *sql.DB
	queries    *db.Queries
	driverName string // "sqlite" or "postgres"
}

// New creates a new repository
func New(database *sql.DB) *Repository {
	return &Repository{
		db:         database,
		queries:    db.New(database),
		driverName: "sqlite", // default
	}
}

// NewWithDB creates a new repository with separate DB and Queries
// This is useful when using a wrapped DB (e.g., for PostgreSQL placeholder conversion)
func NewWithDB(database *sql.DB, queries *db.Queries, driverName string) *Repository {
	return &Repository{
		db:         database,
		queries:    queries,
		driverName: driverName,
	}
}

// NewWithQueries creates a new repository with provided queries
// This is useful for integration tests where you want to use a transaction
func NewWithQueries(queries *db.Queries) *Repository {
	return &Repository{
		db:         nil, // db is nil when using existing queries/transaction
		queries:    queries,
		driverName: "sqlite", // default, will be overridden if needed
	}
}

// SetDriverName sets the driver name (for testing)
func (r *Repository) SetDriverName(driverName string) {
	r.driverName = driverName
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

	// Wrap transaction for PostgreSQL placeholder conversion if needed
	var qtx *db.Queries
	if r.driverName == "postgres" {
		wrappedTx := &postgresTxWrapper{tx: tx}
		qtx = db.New(wrappedTx)
	} else {
		qtx = r.queries.WithTx(tx)
	}

	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit()
}

// postgresTxWrapper wraps sql.Tx and translates SQLite placeholders (?) to PostgreSQL placeholders ($1, $2, etc.)
type postgresTxWrapper struct {
	tx *sql.Tx
}

func (w *postgresTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return w.tx.QueryContext(ctx, convertPlaceholders(query), args...)
}

func (w *postgresTxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return w.tx.QueryRowContext(ctx, convertPlaceholders(query), args...)
}

func (w *postgresTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.tx.ExecContext(ctx, convertPlaceholders(query), args...)
}

func (w *postgresTxWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.tx.PrepareContext(ctx, convertPlaceholders(query))
}

// convertPlaceholders converts SQLite-style ? placeholders to PostgreSQL-style $1, $2, etc.
func convertPlaceholders(query string) string {
	var result strings.Builder
	paramIndex := 1
	inString := false
	var stringChar byte

	for i := 0; i < len(query); i++ {
		ch := query[i]

		// Track if we're inside a string literal
		if ch == '\'' || ch == '"' {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				// Check if it's escaped
				if i > 0 && query[i-1] != '\\' {
					inString = false
				}
			}
		}

		// Only replace ? if we're not inside a string
		if ch == '?' && !inString {
			result.WriteString(fmt.Sprintf("$%d", paramIndex))
			paramIndex++
		} else {
			result.WriteByte(ch)
		}
	}

	return result.String()
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

	// SQLite constraint violations
	if strings.Contains(errMsg, "UNIQUE constraint failed") || strings.Contains(errMsg, "constraint failed") {
		return ErrConflict
	}

	// PostgreSQL constraint violations
	// Error codes: 23505 = unique_violation, 23503 = foreign_key_violation
	if strings.Contains(errMsg, "duplicate key value violates unique constraint") ||
		strings.Contains(errMsg, "23505") ||
		strings.Contains(errMsg, "23503") {
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
	if r.driverName == "postgres" {
		return r.getUserOrdersPostgres(ctx, userID)
	}
	orders, err := r.queries.GetUserOrders(ctx, userID)
	return orders, mapError(err)
}

func (r *Repository) GetOrderWithItems(ctx context.Context, id int64) (db.GetOrderWithItemsRow, error) {
	if r.driverName == "postgres" {
		return r.getOrderWithItemsPostgres(ctx, id)
	}
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

// PostgreSQL-specific implementations

func (r *Repository) getUserOrdersPostgres(ctx context.Context, userID int64) ([]db.GetUserOrdersRow, error) {
	// For now, use a workaround: manually build the result from simpler queries
	// This avoids the json_object vs json_build_object incompatibility
	orders, err := r.queries.ListOrdersByUser(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}

	var result []db.GetUserOrdersRow
	for _, order := range orders {
		items := "[]"
		// Note: In a production implementation, you'd query order_items here
		// For now, we'll return empty items array
		result = append(result, db.GetUserOrdersRow{
			ID:            order.ID,
			UserID:        order.UserID,
			TotalAmount:   order.TotalAmount,
			Status:        order.Status,
			PaymentMethod: order.PaymentMethod,
			CreatedAt:     order.CreatedAt,
			CompletedAt:   order.CompletedAt,
			Items:         items,
		})
	}

	return result, nil
}

func (r *Repository) getOrderWithItemsPostgres(ctx context.Context, id int64) (db.GetOrderWithItemsRow, error) {
	// For now, use a workaround: fetch order and return empty items
	order, err := r.queries.GetOrder(ctx, id)
	if err != nil {
		return db.GetOrderWithItemsRow{}, mapError(err)
	}

	return db.GetOrderWithItemsRow{
		ID:            order.ID,
		UserID:        order.UserID,
		TotalAmount:   order.TotalAmount,
		Status:        order.Status,
		PaymentMethod: order.PaymentMethod,
		CreatedAt:     order.CreatedAt,
		CompletedAt:   order.CompletedAt,
		Items:         "[]",
	}, nil
}
