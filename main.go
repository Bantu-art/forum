package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	handlers "forum/authentication"
	"forum/controllers"
	"forum/utils"
)

func main() {
	if len(os.Args) != 1 {
		log.Fatal("Usage: go run main.go ")
	}
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
	http.HandleFunc("/liked", controllers.LikedPosts)

	http.HandleFunc("/signout", handlers.SignOutHandler(db))

	// Initialize post handler
	postHandler := controllers.NewPostHandler()

	// http.Handle("/post", postHandler)
	http.Handle("/", postHandler) // Handle root for posts

	// Initialize profile handler
	profileHandler := controllers.NewProfileHandler()
	http.Handle("/profile/", profileHandler)

	// Initialize category handler
	categoryHandler := controllers.NewCategoryHandler()
	http.Handle("/categories", categoryHandler)
	http.Handle("/category", categoryHandler)

	// Static file protection middleware
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	protectedStatic := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block direct access completely
		referer := r.Header.Get("Referer")
		if referer == "" || !strings.Contains(referer, r.Host) {
			log.Printf("Blocked direct access attempt to: %s from: %s", r.URL.Path, r.RemoteAddr)
			utils.RenderErrorPage(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
			return
		}

		staticHandler.ServeHTTP(w, r)
	})

	http.Handle("/static/", protectedStatic)
	
	fmt.Println("Server opened at port 8000...http://localhost:8000/")

	// Start server
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
