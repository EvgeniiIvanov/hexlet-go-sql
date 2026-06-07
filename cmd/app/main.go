package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"example.com/go-sql/internal/storage"
)

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
