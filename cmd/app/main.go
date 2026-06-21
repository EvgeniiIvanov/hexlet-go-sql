package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alecthomas/kong"

	"example.com/go-sql/internal/database"
	"example.com/go-sql/internal/storage"
)

var CLI struct {
	DBPath string `help:"Path to SQLite database" default:"data.db"`

	CourseAdd  CourseAddCmd  `cmd:"" help:"Add a new course"`
	CourseList CourseListCmd `cmd:"" help:"List courses"`
	CourseFind CourseFindCmd `cmd:"" help:"Find courses by IDs"`

	UserAdd  UserAddCmd  `cmd:"" help:"Add a new user"`
	UserList UserListCmd `cmd:"" help:"List all users"`
	UserGet  UserGetCmd  `cmd:"" help:"Get user by ID"`
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

	kongCtx := kong.Parse(&CLI,
		kong.Name("gosql"),
		kong.Description("A CLI tool for managing SQLite databases"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Bind(courseRepo),
		kong.Bind(userRepo),
	)

	if err := kongCtx.Run(); err != nil {
		log.Fatalf("run command: %v", err)
	}
}
