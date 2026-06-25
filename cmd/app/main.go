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
	"example.com/go-sql/internal/models"
	"example.com/go-sql/internal/repository"
)

var CLI struct {
	DBPath string `help:"Path to SQLite database" default:"data.db"`

	CourseAdd        CourseAddCmd        `cmd:"" help:"Add a new course"`
	CourseList       CourseListCmd       `cmd:"" help:"List courses"`
	CourseFind       CourseFindCmd       `cmd:"" help:"Find courses by IDs"`
	CourseBulkUpsert CourseBulkUpsertCmd `cmd:"" help:"Bulk upsert courses from JSON file or stdin"`

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
}

// Course commands

type CourseAddCmd struct {
	Slug  string `short:"s" help:"Course slug (unique identifier)" required:""`
	Title string `short:"t" help:"Course title" required:""`
	Price int64  `short:"p" help:"Course price in USD" default:"0"`
}

func (cmd *CourseAddCmd) Run(ctx context.Context, queries *db.Queries) error {
	course, err := queries.CreateCourse(ctx, db.CreateCourseParams{
		Slug:  cmd.Slug,
		Title: cmd.Title,
		Price: cmd.Price,
	})
	if err != nil {
		return fmt.Errorf("create course: %w", err)
	}

	return printJSON(models.FromDBCourse(course))
}

type CourseListCmd struct{}

func (cmd *CourseListCmd) Run(ctx context.Context, queries *db.Queries) error {
	courses, err := queries.ListCourses(ctx)
	if err != nil {
		return fmt.Errorf("list courses: %w", err)
	}

	return printJSON(models.FromDBCourses(courses))
}

type CourseFindCmd struct {
	IDs []int64 `arg:"" help:"Course IDs to find"`
}

func (cmd *CourseFindCmd) Run(ctx context.Context, queries *db.Queries) error {
	courses, err := queries.FindCoursesByIDs(ctx, cmd.IDs)
	if err != nil {
		return fmt.Errorf("find courses: %w", err)
	}

	return printJSON(models.FromDBCourses(courses))
}

type CourseBulkUpsertCmd struct {
	File string `short:"f" help:"JSON file with courses array (use '-' for stdin)" default:"-"`
}

type CourseDTO struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int64  `json:"price"`
}

func (cmd *CourseBulkUpsertCmd) Run(ctx context.Context, repo *repository.Repository) error {
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
	if err := repo.BulkUpsertCourses(ctx, params); err != nil {
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

func (cmd *UserAddCmd) Run(ctx context.Context, queries *db.Queries) error {
	var name sql.NullString
	var age sql.NullInt64

	if cmd.Name != "" {
		name = sql.NullString{String: cmd.Name, Valid: true}
	}
	if cmd.Age > 0 {
		age = sql.NullInt64{Int64: cmd.Age, Valid: true}
	}

	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email: cmd.Email,
		Name:  name,
		Age:   age,
	})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return printJSON(models.FromDBUser(user))
}

type UserListCmd struct{}

func (cmd *UserListCmd) Run(ctx context.Context, queries *db.Queries) error {
	users, err := queries.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	return printJSON(models.FromDBUsers(users))
}

type UserGetCmd struct {
	ID int64 `arg:"" help:"User ID"`
}

func (cmd *UserGetCmd) Run(ctx context.Context, queries *db.Queries) error {
	user, err := queries.GetUser(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	return printJSON(models.FromDBUser(user))
}

type UserBulkUpsertCmd struct {
	File string `short:"f" help:"JSON file with users array (use '-' for stdin)" default:"-"`
}

type UserDTO struct {
	Email string  `json:"email"`
	Name  *string `json:"name,omitempty"`
	Age   *int64  `json:"age,omitempty"`
}

func (cmd *UserBulkUpsertCmd) Run(ctx context.Context, repo *repository.Repository) error {
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
	if err := repo.BulkUpsertUsers(ctx, params); err != nil {
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

func (cmd *EnrollmentCreateCmd) Run(ctx context.Context, repo *repository.Repository) error {
	enrollment, err := repo.EnrollUser(ctx, cmd.UserID, cmd.CourseID)
	if err != nil {
		return fmt.Errorf("enroll user: %w", err)
	}

	return printJSON(models.FromDBEnrollment(enrollment))
}

type EnrollmentCancelCmd struct {
	UserID   int64 `short:"u" help:"User ID" required:""`
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentCancelCmd) Run(ctx context.Context, repo *repository.Repository) error {
	if err := repo.CancelEnrollment(ctx, cmd.UserID, cmd.CourseID); err != nil {
		return fmt.Errorf("cancel enrollment: %w", err)
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

func (cmd *EnrollmentCompleteCmd) Run(ctx context.Context, repo *repository.Repository) error {
	if err := repo.CompleteEnrollment(ctx, cmd.UserID, cmd.CourseID); err != nil {
		return fmt.Errorf("complete enrollment: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully completed enrollment for user %d in course %d", cmd.UserID, cmd.CourseID),
	}

	return printJSON(result)
}

type EnrollmentListCmd struct{}

func (cmd *EnrollmentListCmd) Run(ctx context.Context, queries *db.Queries) error {
	enrollments, err := queries.ListEnrollments(ctx)
	if err != nil {
		return fmt.Errorf("list enrollments: %w", err)
	}

	return printJSON(models.FromDBListEnrollmentsRows(enrollments))
}

type EnrollmentByUserCmd struct {
	UserID int64 `short:"u" help:"User ID" required:""`
}

func (cmd *EnrollmentByUserCmd) Run(ctx context.Context, queries *db.Queries) error {
	enrollments, err := queries.ListEnrollmentsByUser(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("get user enrollments: %w", err)
	}

	return printJSON(models.FromDBListEnrollmentsByUserRows(enrollments))
}

type EnrollmentByCourseCmd struct {
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentByCourseCmd) Run(ctx context.Context, queries *db.Queries) error {
	enrollments, err := queries.ListEnrollmentsByCourse(ctx, cmd.CourseID)
	if err != nil {
		return fmt.Errorf("get course enrollments: %w", err)
	}

	return printJSON(models.FromDBListEnrollmentsByCourseRows(enrollments))
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

	// Create queries and repository
	queries := db.New(dbConn)
	repo := repository.New(dbConn)

	kongCtx := kong.Parse(&CLI,
		kong.Name("gosql"),
		kong.Description("A CLI tool for managing SQLite databases with sqlc"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Bind(queries),
		kong.Bind(repo),
	)

	if err := kongCtx.Run(); err != nil {
		log.Fatalf("run command: %v", err)
	}
}
