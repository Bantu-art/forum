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
	http.HandleFunc("/created", controllers.CreatedPosts)

	http.HandleFunc("/signout", handlers.SignOutHandler(db))
	// Add other route handlers...
	// http.Handle("/", &controllers.PostHandler{})

	postHandler := controllers.NewPostHandler()

	// http.Handle("/post", postHandler)
	http.Handle("/", postHandler) // Handle root for posts

	// Initialize profile handler
	profileHandler := controllers.NewProfileHandler()
	http.Handle("/profile", profileHandler)

	// Initialize category handler
	categoryHandler := controllers.NewCategoryHandler()
	http.Handle("/categories", categoryHandler)
	http.Handle("/category", categoryHandler)

	// Static file serving
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Server opened at port 8000...http://localhost:8000/")

	// Start server
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
