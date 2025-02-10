package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/controllers"
	categories "forum/controllers/categories"
	postcontollers "forum/controllers/post_contollers"
	postfilters "forum/controllers/post_filters"
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
	http.HandleFunc("/created", postfilters.CreatedPosts)
	http.HandleFunc("/liked", postfilters.LikedPosts)

	http.HandleFunc("/signout", handlers.SignOutHandler(db))

	// Initialize post handler
	postHandler := postcontollers.NewPostHandler()

	// http.Handle("/post", postHandler)
	http.Handle("/", postHandler) // Handle root for posts

	// Initialize profile handler
	profileHandler := controllers.NewProfileHandler()
	http.Handle("/profile/", profileHandler)

	// Initialize category handler
	categoryHandler := categories.NewCategoryHandler()
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
