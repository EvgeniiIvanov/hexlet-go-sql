package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/go-sql/internal/database"
	"example.com/go-sql/internal/storage"
)

// Helper functions to create pointers for nullable fields
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	// Create user repository
	userRepo := storage.NewUserRepository(db)

	fmt.Println("=== User Repository Examples ===\n")

	// Example 1: Create user with all fields
	fmt.Println("1. Creating user with name and age...")
	user1, err := userRepo.Create(ctx, storage.CreateUserDTO{
		Email: "john@example.com",
		Name:  stringPtr("John Doe"),
		Age:   intPtr(30),
	})
	if err != nil {
		log.Fatalf("failed to create user1: %v", err)
	}
	printUser(user1)

	// Example 2: Create user with nullable fields as NULL
	fmt.Println("\n2. Creating user with NULL name and age...")
	user2, err := userRepo.Create(ctx, storage.CreateUserDTO{
		Email: "anonymous@example.com",
		Name:  nil, // NULL
		Age:   nil, // NULL
	})
	if err != nil {
		log.Fatalf("failed to create user2: %v", err)
	}
	printUser(user2)

	// Example 3: Create user with only name
	fmt.Println("\n3. Creating user with only name...")
	user3, err := userRepo.Create(ctx, storage.CreateUserDTO{
		Email: "alice@example.com",
		Name:  stringPtr("Alice Smith"),
		Age:   nil, // NULL
	})
	if err != nil {
		log.Fatalf("failed to create user3: %v", err)
	}
	printUser(user3)

	// Example 4: Update user (only update name, keep age)
	fmt.Println("\n4. Updating user1's name only...")
	updated, err := userRepo.Update(ctx, storage.UpdateUserDTO{
		ID:   user1.ID,
		Name: stringPtr("John Updated"),
		Age:  nil, // Don't update age (COALESCE keeps existing value)
	})
	if err != nil {
		log.Fatalf("failed to update user: %v", err)
	}
	printUser(updated)

	// Example 5: List all users
	fmt.Println("\n5. Listing all users...")
	users, err := userRepo.List(ctx)
	if err != nil {
		log.Fatalf("failed to list users: %v", err)
	}
	for i, u := range users {
		fmt.Printf("  User %d: ", i+1)
		printUser(u)
	}

	// Example 6: Find by email
	fmt.Println("\n6. Finding user by email...")
	found, err := userRepo.FindByEmail(ctx, "alice@example.com")
	if err != nil {
		log.Fatalf("failed to find user: %v", err)
	}
	if found != nil {
		printUser(*found)
	}

	// Example 7: Get specific user
	fmt.Println("\n7. Getting user by ID...")
	fetched, err := userRepo.Get(ctx, user1.ID)
	if err != nil {
		log.Fatalf("failed to get user: %v", err)
	}
	printUser(fetched)

	fmt.Println("\n=== Examples Complete ===")
}

func printUser(u storage.User) {
	name := "NULL"
	if u.Name != nil {
		name = *u.Name
	}

	age := "NULL"
	if u.Age != nil {
		age = fmt.Sprintf("%d", *u.Age)
	}

	fmt.Printf("ID: %d | Email: %s | Name: %s | Age: %s | Created: %s\n",
		u.ID, u.Email, name, age, u.CreatedAt)
}
