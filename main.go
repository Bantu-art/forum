package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/handlers"
	"forum/utils"
)

func main() {
	// Initialize database
	db, err := utils.InitialiseDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// Initialize handlers with database
	handlers.InitDB(db)

	// Setup routes
	http.HandleFunc("/signup", handlers.SignUpHandler)
	http.HandleFunc("/signin", handlers.SignInHandler)
	// Add other route handlers...

	// Start server
	port := ":8080"
	fmt.Printf("Server starting on %s...\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
