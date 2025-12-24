package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Generate hash for "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		os.Exit(1)
	}
	hashStr := string(hash)
	fmt.Println("Generated hash:", hashStr)

	// Verify the hash works
	if err := bcrypt.CompareHashAndPassword([]byte(hashStr), []byte("admin123")); err != nil {
		fmt.Println("Hash verification failed:", err)
		os.Exit(1)
	}
	fmt.Println("Hash verified successfully!")

	// Connect to database and update
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://swiftlet:swiftlet@localhost:5432/swiftlet?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.ExecContext(context.Background(),
		"UPDATE users SET password_hash = $1 WHERE email = 'admin@swiftlead.id'",
		hashStr)
	if err != nil {
		fmt.Println("Error updating password:", err)
		os.Exit(1)
	}

	fmt.Println("Password updated successfully!")
	fmt.Println("Login with: admin@swiftlead.id / admin123")
}
