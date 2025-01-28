package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

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

	//start session cleanup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure the go routine is stoppend when the application exists

	//set interval  and start the cleanup session
	cleanUpInterval := 1 * time.Second
	utils.StartSessionsCLeanUp(ctx, db, cleanUpInterval)

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
