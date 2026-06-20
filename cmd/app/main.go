package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	_ "modernc.org/sqlite"

	"example.com/go-sql/internal/storage"
)

var CLI struct {
	DBPath string `help:"Path to SQLite database" default:"data.db"`

	CourseAdd  CourseAddCmd  `cmd:"" help:"Add a new course"`
	CourseList CourseListCmd `cmd:"" help:"List courses"`
	CourseFind CourseFindCmd `cmd:"" help:"Find courses by IDs"`
}

type CourseAddCmd struct {
	Slug  string `short:"s" help:"Course slug (unique identifier)" required:""`
	Title string `short:"t" help:"Course title" required:""`
	Price int    `short:"p" help:"Course price in USD" default:"0"`
}

func (cmd *CourseAddCmd) Run(ctx context.Context, db *sql.DB) error {
	course := storage.Course{
		Slug:  cmd.Slug,
		Title: cmd.Title,
		Price: cmd.Price,
	}

	id, err := CreateCourse(ctx, db, course)
	if err != nil {
		return fmt.Errorf("create course: %w", err)
	}

	fmt.Printf("Course created successfully! ID: %d\n", id)
	return nil
}

type CourseListCmd struct {
	Limit  int    `short:"l" help:"Number of courses to return" default:"10"`
	Offset int    `short:"o" help:"Offset for pagination" default:"0"`
	Order  string `short:"r" help:"Order by field (id, slug, title, price)" default:"id" enum:"id,slug,title,price"`
}

func (cmd *CourseListCmd) Run(ctx context.Context, db *sql.DB) error {
	courses, err := ListCourses(ctx, db, cmd.Limit, cmd.Offset, cmd.Order)
	if err != nil {
		return fmt.Errorf("list courses: %w", err)
	}

	if len(courses) == 0 {
		fmt.Println("No courses found")
		return nil
	}

	fmt.Printf("Found %d courses:\n\n", len(courses))
	for _, c := range courses {
		fmt.Printf("  ID: %d | Slug: %s | Title: %s | Price: $%d\n",
			c.ID, c.Slug, c.Title, c.Price)
	}
	return nil
}

type CourseFindCmd struct {
	IDs []int64 `arg:"" name:"ids" help:"Course IDs to find" required:""`
}

func (cmd *CourseFindCmd) Run(ctx context.Context, db *sql.DB) error {
	courses, err := FindCoursesByIDs(ctx, db, cmd.IDs)
	if err != nil {
		return fmt.Errorf("find courses: %w", err)
	}

	if len(courses) == 0 {
		fmt.Println("No courses found for given IDs")
		return nil
	}

	fmt.Printf("Found %d courses:\n\n", len(courses))
	for _, c := range courses {
		fmt.Printf("  ID: %d | Slug: %s | Title: %s | Price: $%d\n",
			c.ID, c.Slug, c.Title, c.Price)
	}
	return nil
}

func CreateCourse(ctx context.Context, db *sql.DB, c storage.Course) (int64, error) {
	const query = `INSERT INTO courses (slug, title, price) VALUES (?, ?, ?) RETURNING id`

	var id int64
	if err := db.QueryRowContext(ctx, query, c.Slug, c.Title, c.Price).Scan(&id); err != nil {
		return 0, fmt.Errorf("create course: %w", err)
	}
	return id, nil
}

var allowedOrder = map[string]string{
	"price_asc":  "price ASC",
	"price_desc": "price DESC",
	"title_asc":  "title ASC",
}

func ListCourses(ctx context.Context, db *sql.DB, limit, offset int, order string) ([]storage.Course, error) {
	ord, ok := allowedOrder[order]
	if !ok {
		ord = "id ASC"
	}

	query := `
        SELECT id, slug, title, price
        FROM courses
        ORDER BY ` + ord + `
        LIMIT ? OFFSET ?
    `

	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	var courses []storage.Course
	for rows.Next() {
		var c storage.Course
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func FindCoursesByIDs(ctx context.Context, db *sql.DB, ids []int64) ([]storage.Course, error) {
	if len(ids) == 0 {
		return []storage.Course{}, nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	query := `
		SELECT id, slug, title, price
		FROM courses
		WHERE id IN (` + placeholders + `)
	`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find courses: %w", err)
	}
	defer rows.Close()

	var courses []storage.Course
	for rows.Next() {
		var c storage.Course
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.Price); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	db, err := sql.Open("sqlite", "file:data.db?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	// Define pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(30 * time.Second)

	const schemaCourses = `CREATE TABLE IF NOT EXISTS courses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		price INTEGER NOT NULL DEFAULT 0
	)`

	if _, err := db.ExecContext(ctx, schemaCourses); err != nil {
		log.Fatalf("create table: %v", err)
	}

	kongCtx := kong.Parse(&CLI,
		kong.Name("gosql"),
		kong.Description("A CLI tool for managing SQLite database of courses"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Bind(db),
	)

	if err := kongCtx.Run(); err != nil {
		log.Fatalf("run command: %v", err)
	}
}
