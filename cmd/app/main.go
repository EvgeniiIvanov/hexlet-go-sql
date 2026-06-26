package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/alecthomas/kong"

	"example.com/go-sql/internal/database"
	"example.com/go-sql/internal/db"
	"example.com/go-sql/internal/repository"
	"example.com/go-sql/internal/service"
)

var CLI struct {
	DBPath string `help:"Path to SQLite database" default:"data.db"`

	CourseAdd             CourseAddCmd             `cmd:"" help:"Add a new course"`
	CourseList            CourseListCmd            `cmd:"" help:"List courses"`
	CourseFind            CourseFindCmd            `cmd:"" help:"Find courses by IDs"`
	CourseWithEnrollments CourseWithEnrollmentsCmd `cmd:"" help:"Get a course with all its enrollments"`
	CourseBulkUpsert      CourseBulkUpsertCmd      `cmd:"" help:"Bulk upsert courses from JSON file or stdin"`

	UserAdd        UserAddCmd        `cmd:"" help:"Add a new user"`
	UserList       UserListCmd       `cmd:"" help:"List all users"`
	UserGet        UserGetCmd        `cmd:"" help:"Get user by ID"`
	UserBulkUpsert UserBulkUpsertCmd `cmd:"" help:"Bulk upsert users from JSON file or stdin"`

	EnrollmentCreate   EnrollmentCreateCmd   `cmd:"" help:"Enroll a user in a course (using transaction)"`
	EnrollmentCancel   EnrollmentCancelCmd   `cmd:"" help:"Cancel a user's enrollment"`
	EnrollmentComplete EnrollmentCompleteCmd `cmd:"" help:"Mark an enrollment as completed"`
	EnrollmentList     EnrollmentListCmd     `cmd:"" help:"List all enrollments"`
	EnrollmentByUser   EnrollmentByUserCmd   `cmd:"" help:"List enrollments for a specific user"`
	EnrollmentByCourse EnrollmentByCourseCmd `cmd:"" help:"List enrollments for a specific course"`

	OrderCreate       OrderCreateCmd       `cmd:"" help:"Create an order and enroll user in courses"`
	OrderGet          OrderGetCmd          `cmd:"" help:"Get order details with items"`
	OrderList         OrderListCmd         `cmd:"" help:"List all orders"`
	OrdersByUser      OrdersByUserCmd      `cmd:"" help:"Get user's order history"`
	OrderRefund       OrderRefundCmd       `cmd:"" help:"Refund an order"`
	CheckCourseAccess CheckCourseAccessCmd `cmd:"" help:"Check if user owns a course"`
}

// Course commands

type CourseAddCmd struct {
	Slug  string `short:"s" help:"Course slug (unique identifier)" required:""`
	Title string `short:"t" help:"Course title" required:""`
	Price int64  `short:"p" help:"Course price in USD" default:"0"`
}

func (cmd *CourseAddCmd) Run(ctx context.Context, svc *service.Service) error {
	course, err := svc.CreateCourse(ctx, cmd.Slug, cmd.Title, cmd.Price)
	if err != nil {
		return err
	}

	return printJSON(course)
}

type CourseListCmd struct {
	Limit  int64 `short:"l" help:"Number of courses to return" default:"10"`
	Offset int64 `short:"o" help:"Offset for pagination" default:"0"`
}

func (cmd *CourseListCmd) Run(ctx context.Context, svc *service.Service) error {
	courses, err := svc.ListCourses(ctx, cmd.Limit, cmd.Offset)
	if err != nil {
		return err
	}

	return printJSON(courses)
}

type CourseFindCmd struct {
	IDs []int64 `arg:"" help:"Course IDs to find"`
}

func (cmd *CourseFindCmd) Run(ctx context.Context, svc *service.Service) error {
	courses, err := svc.FindCoursesByIDs(ctx, cmd.IDs)
	if err != nil {
		return err
	}

	return printJSON(courses)
}

type CourseWithEnrollmentsCmd struct {
	ID int64 `arg:"" help:"Course ID"`
}

func (cmd *CourseWithEnrollmentsCmd) Run(ctx context.Context, svc *service.Service) error {
	course, err := svc.GetCourseWithEnrollments(ctx, cmd.ID)
	if err != nil {
		return err
	}

	return printJSON(course)
}

type CourseBulkUpsertCmd struct {
	File string `short:"f" help:"JSON file with courses array (use '-' for stdin)" default:"-"`
}

type CourseDTO struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int64  `json:"price"`
}

func (cmd *CourseBulkUpsertCmd) Run(ctx context.Context, svc *service.Service) error {
	start := time.Now()

	var dtos []CourseDTO
	var data []byte
	var err error

	if cmd.File == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(cmd.File)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
	}

	if err := json.Unmarshal(data, &dtos); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	params := make([]db.UpsertCourseParams, len(dtos))
	for i, dto := range dtos {
		params[i] = db.UpsertCourseParams{
			Slug:  dto.Slug,
			Title: dto.Title,
			Price: dto.Price,
		}
	}

	operationStart := time.Now()
	if err := svc.BulkUpsertCourses(ctx, params); err != nil {
		return fmt.Errorf("bulk upsert: %w", err)
	}
	operationDuration := time.Since(operationStart)

	result := map[string]interface{}{
		"success":        true,
		"count":          len(dtos),
		"message":        fmt.Sprintf("Successfully upserted %d courses", len(dtos)),
		"operation_time": operationDuration.String(),
		"total_time":     time.Since(start).String(),
		"avg_per_record": (operationDuration / time.Duration(len(dtos))).String(),
	}

	return printJSON(result)
}

// User commands

type UserAddCmd struct {
	Email string `short:"e" help:"User email" required:""`
	Name  string `short:"n" help:"User name (optional)"`
	Age   int64  `short:"a" help:"User age (optional)"`
}

func (cmd *UserAddCmd) Run(ctx context.Context, svc *service.Service) error {
	var name sql.NullString
	var age sql.NullInt64

	if cmd.Name != "" {
		name = sql.NullString{String: cmd.Name, Valid: true}
	}
	if cmd.Age > 0 {
		age = sql.NullInt64{Int64: cmd.Age, Valid: true}
	}

	user, err := svc.CreateUser(ctx, cmd.Email, name, age)
	if err != nil {
		return err
	}

	return printJSON(user)
}

type UserListCmd struct {
	Limit  int64 `short:"l" help:"Number of users to return" default:"10"`
	Offset int64 `short:"o" help:"Offset for pagination" default:"0"`
}

func (cmd *UserListCmd) Run(ctx context.Context, svc *service.Service) error {
	users, err := svc.ListUsers(ctx, cmd.Limit, cmd.Offset)
	if err != nil {
		return err
	}

	return printJSON(users)
}

type UserGetCmd struct {
	ID int64 `arg:"" help:"User ID"`
}

func (cmd *UserGetCmd) Run(ctx context.Context, svc *service.Service) error {
	user, err := svc.GetUser(ctx, cmd.ID)
	if err != nil {
		return err
	}

	return printJSON(user)
}

type UserBulkUpsertCmd struct {
	File string `short:"f" help:"JSON file with users array (use '-' for stdin)" default:"-"`
}

type UserDTO struct {
	Email string  `json:"email"`
	Name  *string `json:"name,omitempty"`
	Age   *int64  `json:"age,omitempty"`
}

func (cmd *UserBulkUpsertCmd) Run(ctx context.Context, svc *service.Service) error {
	start := time.Now()

	var dtos []UserDTO
	var data []byte
	var err error

	if cmd.File == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(cmd.File)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
	}

	if err := json.Unmarshal(data, &dtos); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	params := make([]db.UpsertUserParams, len(dtos))
	for i, dto := range dtos {
		params[i] = db.UpsertUserParams{
			Email: dto.Email,
		}
		if dto.Name != nil {
			params[i].Name = sql.NullString{String: *dto.Name, Valid: true}
		}
		if dto.Age != nil {
			params[i].Age = sql.NullInt64{Int64: *dto.Age, Valid: true}
		}
	}

	operationStart := time.Now()
	if err := svc.BulkUpsertUsers(ctx, params); err != nil {
		return fmt.Errorf("bulk upsert: %w", err)
	}
	operationDuration := time.Since(operationStart)

	result := map[string]interface{}{
		"success":        true,
		"count":          len(dtos),
		"message":        fmt.Sprintf("Successfully upserted %d users", len(dtos)),
		"operation_time": operationDuration.String(),
		"total_time":     time.Since(start).String(),
		"avg_per_record": (operationDuration / time.Duration(len(dtos))).String(),
	}

	return printJSON(result)
}

// Enrollment commands

type EnrollmentCreateCmd struct {
	UserID   int64 `short:"u" help:"User ID" required:""`
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentCreateCmd) Run(ctx context.Context, svc *service.Service) error {
	enrollment, err := svc.EnrollUser(ctx, cmd.UserID, cmd.CourseID)
	if err != nil {
		return err
	}

	return printJSON(enrollment)
}

type EnrollmentCancelCmd struct {
	UserID   int64 `short:"u" help:"User ID" required:""`
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentCancelCmd) Run(ctx context.Context, svc *service.Service) error {
	if err := svc.CancelEnrollment(ctx, cmd.UserID, cmd.CourseID); err != nil {
		return err
	}

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully cancelled enrollment for user %d in course %d", cmd.UserID, cmd.CourseID),
	}

	return printJSON(result)
}

type EnrollmentCompleteCmd struct {
	UserID   int64 `short:"u" help:"User ID" required:""`
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentCompleteCmd) Run(ctx context.Context, svc *service.Service) error {
	if err := svc.CompleteEnrollment(ctx, cmd.UserID, cmd.CourseID); err != nil {
		return err
	}

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully completed enrollment for user %d in course %d", cmd.UserID, cmd.CourseID),
	}

	return printJSON(result)
}

type EnrollmentListCmd struct {
	Limit  int64 `short:"l" help:"Number of enrollments to return" default:"10"`
	Offset int64 `short:"o" help:"Offset for pagination" default:"0"`
}

func (cmd *EnrollmentListCmd) Run(ctx context.Context, svc *service.Service) error {
	enrollments, err := svc.ListEnrollments(ctx, cmd.Limit, cmd.Offset)
	if err != nil {
		return err
	}

	return printJSON(enrollments)
}

type EnrollmentByUserCmd struct {
	UserID int64 `short:"u" help:"User ID" required:""`
}

func (cmd *EnrollmentByUserCmd) Run(ctx context.Context, svc *service.Service) error {
	enrollments, err := svc.ListEnrollmentsByUser(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	return printJSON(enrollments)
}

type EnrollmentByCourseCmd struct {
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentByCourseCmd) Run(ctx context.Context, svc *service.Service) error {
	enrollments, err := svc.ListEnrollmentsByCourse(ctx, cmd.CourseID)
	if err != nil {
		return err
	}

	return printJSON(enrollments)
}

// Order commands

type OrderCreateCmd struct {
	UserID        int64   `short:"u" help:"User ID" required:""`
	CourseIDs     []int64 `short:"c" help:"Course IDs to purchase" required:""`
	PaymentMethod string  `short:"p" help:"Payment method (card, paypal, etc.)" default:"card"`
}

func (cmd *OrderCreateCmd) Run(ctx context.Context, svc *service.Service) error {
	order, err := svc.CreateOrder(ctx, cmd.UserID, cmd.CourseIDs, cmd.PaymentMethod)
	if err != nil {
		return err
	}

	return printJSON(order)
}

type OrderGetCmd struct {
	ID int64 `arg:"" help:"Order ID"`
}

func (cmd *OrderGetCmd) Run(ctx context.Context, svc *service.Service) error {
	order, err := svc.GetOrderWithItems(ctx, cmd.ID)
	if err != nil {
		return err
	}

	return printJSON(order)
}

type OrderListCmd struct {
	Limit  int64 `short:"l" help:"Number of orders to return" default:"10"`
	Offset int64 `short:"o" help:"Offset for pagination" default:"0"`
}

func (cmd *OrderListCmd) Run(ctx context.Context, svc *service.Service) error {
	orders, err := svc.ListOrders(ctx, cmd.Limit, cmd.Offset)
	if err != nil {
		return err
	}

	return printJSON(orders)
}

type OrdersByUserCmd struct {
	UserID int64 `short:"u" help:"User ID" required:""`
}

func (cmd *OrdersByUserCmd) Run(ctx context.Context, svc *service.Service) error {
	orders, err := svc.GetUserOrders(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	return printJSON(orders)
}

type OrderRefundCmd struct {
	ID int64 `arg:"" help:"Order ID to refund"`
}

func (cmd *OrderRefundCmd) Run(ctx context.Context, svc *service.Service) error {
	if err := svc.RefundOrder(ctx, cmd.ID); err != nil {
		return err
	}

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Order %d has been refunded", cmd.ID),
	}

	return printJSON(result)
}

type CheckCourseAccessCmd struct {
	UserID   int64 `short:"u" help:"User ID" required:""`
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *CheckCourseAccessCmd) Run(ctx context.Context, svc *service.Service) error {
	owns, err := svc.CheckUserOwnsCourse(ctx, cmd.UserID, cmd.CourseID)
	if err != nil {
		return err
	}

	result := map[string]interface{}{
		"user_id":    cmd.UserID,
		"course_id":  cmd.CourseID,
		"has_access": owns,
	}

	return printJSON(result)
}

// Helper functions

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Connect to database
	dbConn, err := database.Connect(ctx, database.DefaultConfig())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer dbConn.Close()

	// Initialize schema
	if err := database.InitSchema(ctx, dbConn); err != nil {
		log.Fatalf("schema initialization failed: %v", err)
	}

	// Create repository and service
	repo := repository.New(dbConn)
	svc := service.New(repo)

	kongCtx := kong.Parse(&CLI,
		kong.Name("gosql"),
		kong.Description("A CLI tool for managing SQLite databases with sqlc"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Bind(svc),
	)

	if err := kongCtx.Run(); err != nil {
		log.Fatalf("run command: %v", err)
	}
}
