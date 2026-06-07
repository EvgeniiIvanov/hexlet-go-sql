package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
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

	const schema = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR(255) NOT NULL UNIQUE,
		name VARCHAR(255)
	)`

	if _, err := db.ExecContext(ctx, schema); err != nil {
		log.Fatalf("create table: %v", err)
	}

	const insert = `INSERT INTO users (email, name) VALUES (?, ?) ON CONFLICT DO NOTHING;`

	if _, err := db.ExecContext(ctx, insert, "rick@cartoon.com", "Rick"); err != nil {
		log.Fatalf("insert: %v", err)
	}

	if _, err := db.ExecContext(ctx, insert, "morty@cartoon.com", "Morty"); err != nil {
		log.Fatalf("insert: %v", err)
	}

	if _, err := db.ExecContext(ctx, insert, "summer@cartoon.com", "Summer"); err != nil {
		log.Fatalf("insert: %v", err)
	}

	if _, err := db.ExecContext(ctx, insert, "harry@movie.com", "Harry"); err != nil {
		log.Fatalf("insert: %v", err)
	}

	if _, err := db.ExecContext(ctx, insert, "hermione@movie.com", "Hermione"); err != nil {
		log.Fatalf("insert: %v", err)
	}

	const update = `UPDATE users SET name = ? WHERE email = ?`
	if _, err := db.ExecContext(ctx, update, "Morty Smith", "morty@cartoon.com"); err != nil {
		log.Fatalf("update: %v", err)
	}

	var u User
	err = db.QueryRowContext(ctx,
		`SELECT id, email, name FROM users WHERE email = ?`,
		"rick@cartoon.com",
	).Scan(&u.ID, &u.Email, &u.Name)
	if err != nil {
		log.Fatalf("query: %v", err)
	}

	var characters []User
	rows, err := db.QueryContext(ctx, `SELECT id, email, name FROM users ORDER BY name`)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name); err != nil {
			log.Fatalf("scan: %v", err)
		}
		characters = append(characters, u)
	}

	payload, _ := json.MarshalIndent(u, "", "  ")
	log.Printf("loaded user: %s", payload)

	charactersPayload, _ := json.MarshalIndent(characters, "", "  ")
	log.Printf("loaded users: %s", charactersPayload)
}
