package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/controllers"
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
	utils.InitSessionManager(utils.GlobalDB)
	// Setup routes
	http.HandleFunc("/signup", handlers.SignUpHandler)
	http.HandleFunc("/signin", handlers.SignInHandler)

	// http.Handle("/", &controllers.PostHandler{})
	postHandler := controllers.NewPostHandler()
	http.Handle("/", postHandler)
    http.HandleFunc("/react", postHandler.HandleReaction) // Changed to HandleFunc


	//http.HandleFunc("/react", postHandler.HandleReaction)	// Add other route handlers...

	profileHandler := controllers.NewProfileHandler()
	http.Handle("/profile", profileHandler)

	// http.Handle("/post", postHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Server opened at port 3000...http://localhost:8000/")

	// Start server

	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
