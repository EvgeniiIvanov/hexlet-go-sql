package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/alecthomas/kong"

	"example.com/go-sql/internal/database"
	"example.com/go-sql/internal/storage"
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

type CourseAddCmd struct {
	Slug  string `short:"s" help:"Course slug (unique identifier)" required:""`
	Title string `short:"t" help:"Course title" required:""`
	Price int    `short:"p" help:"Course price in USD" default:"0"`
}

func (cmd *CourseAddCmd) Run(ctx context.Context, repo *storage.CourseRepository) error {
	dto := storage.CreateCourseDTO{
		Slug:  cmd.Slug,
		Title: cmd.Title,
		Price: cmd.Price,
	}

	course, err := repo.Create(ctx, dto)
	if err != nil {
		return fmt.Errorf("create course: %w", err)
	}

	return printJSON(course)
}

type CourseListCmd struct {
	Limit  int    `short:"l" help:"Number of courses to return" default:"10"`
	Offset int    `short:"o" help:"Offset for pagination" default:"0"`
	Order  string `short:"r" help:"Order by field (id, slug, title, price)" default:"id" enum:"id,slug,title,price"`
}

func (cmd *CourseListCmd) Run(ctx context.Context, repo *storage.CourseRepository) error {
	courses, err := repo.List(ctx, cmd.Limit, cmd.Offset, cmd.Order)
	if err != nil {
		return fmt.Errorf("list courses: %w", err)
	}

	return printJSON(courses)
}

type CourseFindCmd struct {
	IDs []int64 `arg:"" name:"ids" help:"Course IDs to find" required:""`
}

func (cmd *CourseFindCmd) Run(ctx context.Context, repo *storage.CourseRepository) error {
	courses, err := repo.FindByIDs(ctx, cmd.IDs)
	if err != nil {
		return fmt.Errorf("find courses: %w", err)
	}

	return printJSON(courses)
}

type CourseBulkUpsertCmd struct {
	File string `short:"f" help:"JSON file with courses array (use '-' for stdin)" default:"-"`
}

func (cmd *CourseBulkUpsertCmd) Run(ctx context.Context, repo *storage.CourseRepository) error {
	start := time.Now()

	var dtos []storage.CreateCourseDTO

	// Read from file or stdin
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

	// Parse JSON
	if err := json.Unmarshal(data, &dtos); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	// Perform bulk upsert
	operationStart := time.Now()
	if err := repo.BulkUpsert(ctx, dtos); err != nil {
		return fmt.Errorf("bulk upsert: %w", err)
	}
	operationDuration := time.Since(operationStart)

	// Return success message with timing
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
	Email string  `short:"e" help:"User email (required)" required:""`
	Name  *string `short:"n" help:"User name (optional)"`
	Age   *int    `short:"a" help:"User age (optional)"`
}

func (cmd *UserAddCmd) Run(ctx context.Context, repo *storage.UserRepository) error {
	dto := storage.CreateUserDTO{
		Email: cmd.Email,
		Name:  cmd.Name,
		Age:   cmd.Age,
	}

	user, err := repo.Create(ctx, dto)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return printJSON(user)
}

type UserListCmd struct{}

func (cmd *UserListCmd) Run(ctx context.Context, repo *storage.UserRepository) error {
	users, err := repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	return printJSON(users)
}

type UserGetCmd struct {
	ID int64 `arg:"" help:"User ID to retrieve" required:""`
}

func (cmd *UserGetCmd) Run(ctx context.Context, repo *storage.UserRepository) error {
	user, err := repo.Get(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	return printJSON(user)
}

type UserBulkUpsertCmd struct {
	File string `short:"f" help:"JSON file with users array (use '-' for stdin)" default:"-"`
}

func (cmd *UserBulkUpsertCmd) Run(ctx context.Context, repo *storage.UserRepository) error {
	start := time.Now()

	var dtos []storage.CreateUserDTO

	// Read from file or stdin
	var data []byte
	var err error

	if cmd.File == "-" {
		// Read from stdin
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	} else {
		// Read from file
		data, err = os.ReadFile(cmd.File)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
	}

	// Parse JSON
	if err := json.Unmarshal(data, &dtos); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	// Perform bulk upsert
	operationStart := time.Now()
	if err := repo.BulkUpsert(ctx, dtos); err != nil {
		return fmt.Errorf("bulk upsert: %w", err)
	}
	operationDuration := time.Since(operationStart)

	// Return success message with timing
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

func (cmd *EnrollmentCreateCmd) Run(ctx context.Context, repo *storage.EnrollmentRepository) error {
	enrollment, err := repo.EnrollUser(ctx, cmd.UserID, cmd.CourseID)
	if err != nil {
		return fmt.Errorf("enroll user: %w", err)
	}

	return printJSON(enrollment)
}

type EnrollmentCancelCmd struct {
	UserID   int64 `short:"u" help:"User ID" required:""`
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentCancelCmd) Run(ctx context.Context, repo *storage.EnrollmentRepository) error {
	if err := repo.UnenrollUser(ctx, cmd.UserID, cmd.CourseID); err != nil {
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

func (cmd *EnrollmentCompleteCmd) Run(ctx context.Context, repo *storage.EnrollmentRepository) error {
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

func (cmd *EnrollmentListCmd) Run(ctx context.Context, repo *storage.EnrollmentRepository) error {
	enrollments, err := repo.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list enrollments: %w", err)
	}

	return printJSON(enrollments)
}

type EnrollmentByUserCmd struct {
	UserID int64 `short:"u" help:"User ID" required:""`
}

func (cmd *EnrollmentByUserCmd) Run(ctx context.Context, repo *storage.EnrollmentRepository) error {
	enrollments, err := repo.GetUserEnrollments(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("get user enrollments: %w", err)
	}

	return printJSON(enrollments)
}

type EnrollmentByCourseCmd struct {
	CourseID int64 `short:"c" help:"Course ID" required:""`
}

func (cmd *EnrollmentByCourseCmd) Run(ctx context.Context, repo *storage.EnrollmentRepository) error {
	enrollments, err := repo.GetCourseEnrollments(ctx, cmd.CourseID)
	if err != nil {
		return fmt.Errorf("get course enrollments: %w", err)
	}

	return printJSON(enrollments)
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
	db, err := database.Connect(ctx, database.DefaultConfig())
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := database.InitSchema(ctx, db); err != nil {
		log.Fatalf("schema initialization failed: %v", err)
	}

	// Create repositories
	courseRepo := storage.NewCourseRepository(db)
	userRepo := storage.NewUserRepository(db)
	enrollmentRepo := storage.NewEnrollmentRepository(db)

	kongCtx := kong.Parse(&CLI,
		kong.Name("gosql"),
		kong.Description("A CLI tool for managing SQLite databases"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Bind(courseRepo),
		kong.Bind(userRepo),
		kong.Bind(enrollmentRepo),
	)

	if err := kongCtx.Run(); err != nil {
		log.Fatalf("run command: %v", err)
	}
}
