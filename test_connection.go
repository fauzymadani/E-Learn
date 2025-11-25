package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Test 1: Original format
	dsn1 := "host=localhost port=5432 user=postgres password= dbname=elearning sslmode=disable"
	testConnection("Test 1 (key=value format)", dsn1)

	// Test 2: URL format
	dsn2 := "postgresql://postgres@localhost:5432/elearning?sslmode=disable"
	testConnection("Test 2 (URL format)", dsn2)

	// Test 3: URL format with explicit database
	dsn3 := "postgres://postgres@localhost/elearning?sslmode=disable"
	testConnection("Test 3 (postgres:// URL)", dsn3)
}

func testConnection(name, dsn string) {
	log.Printf("\n=== %s ===", name)
	log.Printf("DSN: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("ERROR Ping: %v\n", err)
		return
	}

	var currentDB string
	err = db.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err != nil {
		log.Printf("ERROR Query: %v\n", err)
		return
	}

	fmt.Printf("✓ Connected to database: %s\n", currentDB)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Printf("User count error: %v", err)
	} else {
		fmt.Printf("✓ User count: %d\n", count)
	}
}
