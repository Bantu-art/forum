package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"forum/utils"
)

type PostHandler struct {
	imageHandler *ImageHandler
}

func NewPostHandler() *PostHandler {
	return &PostHandler{
		imageHandler: NewImageHandler(), // Initialize ImageHandler
	}
}

// Update handler signatures to match http.HandlerFunc
func (ph *PostHandler) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		// Define a custom type for context keys
		type contextKey string

		// Store userID in request context
		ctx := context.WithValue(r.Context(), contextKey("userID"), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Update ServeHTTP method
func (ph *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/create":
		switch r.Method {
		case http.MethodGet:
			ph.authMiddleware(ph.displayCreateForm).ServeHTTP(w, r)
		case http.MethodPost:
			ph.authMiddleware(ph.handleCreatePost).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/":
		if r.Method == http.MethodGet {
			if r.URL.Query().Get("id") != "" {
				ph.handleSinglePost(w, r) // Public access allowed
			} else {
				ph.handleGetPosts(w, r) // Public access allowed
			}
		}
	default:
		http.NotFound(w, r)
	}
}

func (ph *PostHandler) displayCreateForm(w http.ResponseWriter, r *http.Request) {
	categories, err := ph.getAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Error loading categories", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/createpost.html")
	if err != nil {
		log.Printf("Error parsing create form template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Categories []utils.Category
	}{
		Categories: categories,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing create form template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}
func (ph *PostHandler) getAllCategories() ([]utils.Category, error) {
	rows, err := utils.GlobalDB.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []utils.Category
	for rows.Next() {
		var category utils.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			log.Printf("Error scanning category: %v", err)
			continue
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (ph *PostHandler) handleGetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := ph.getAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pageData := utils.PageData{
		IsLoggedIn: ph.checkAuthStatus(r),
		Posts:      posts,
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, pageData); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

func (ph *PostHandler) getAllPosts() ([]utils.Post, error) {
	rows, err := utils.GlobalDB.Query(`
        SELECT p.id, p.user_id, p.title, p.content, p.imagepath, 
               p.post_at, p.likes, p.dislikes, p.comments,
               u.username, u.profile_pic, c.id AS category_id, c.name AS category_name
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN post_categories pc ON p.id = pc.post_id
        LEFT JOIN categories c ON pc.category_id = c.id
        ORDER BY p.post_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []utils.Post
	for rows.Next() {
		var post utils.Post
		var postTime time.Time
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&postTime,
			&post.Likes,
			&post.Dislikes,
			&post.Comments,
			&post.Username,
			&post.ProfilePic,
			&post.CategoryID,
			&post.CategoryName,
		); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		post.PostTime = FormatTimeAgo(postTime)
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

// Update handleCreatePost method
func (ph *PostHandler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error processing form", http.StatusBadRequest)
		return
	}

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")
	categoryID := r.FormValue("category")

	if title == "" || content == "" || categoryID == "" {
		log.Printf("Title, content, and category are required")
		http.Error(w, "Title, content, and category are required", http.StatusBadRequest)
		return
	}
	// Handle image upload
	var imagePath string
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imagePath, err = ph.imageHandler.ProcessImage(file, header)
		if err != nil {
			log.Printf("Error processing image: %v", err)
			// Continue without image if there's an error
		}
	}

	// Prepare the insert statement with image support
	stmt, err := utils.GlobalDB.Prepare(`
        INSERT INTO posts (user_id, title, content, imagepath, post_at, likes, dislikes, comments) 
        VALUES (?, ?, ?, ?, ?, 0, 0, 0)
    `)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Execute the insert
	currentTime := time.Now()
	result, err := stmt.Exec(userID, title, content, imagePath, currentTime)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()

	if categoryID != "" {
		_, err = utils.GlobalDB.Exec(`
            INSERT INTO post_categories (post_id, category_id) 
            VALUES (?, ?)
        `, postID, categoryID)
		if err != nil {
			log.Printf("Error inserting into post_categories: %v", err)
		}
	}
	log.Printf("Successfully created post with ID: %d", postID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func FormatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func (ph *PostHandler) handleSinglePost(w http.ResponseWriter, r *http.Request) {
	// Get post ID from URL query parameter
	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get post from database
	post, err := ph.getPostByID(postID)
	if err != nil {
		log.Printf("Error fetching post: %v", err)
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		return
	}

	// If post not found
	if post == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Parse and execute template
	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, post); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

// Add this helper method to fetch a single post
func (ph *PostHandler) getPostByID(id int64) (*utils.Post, error) {
	row := utils.GlobalDB.QueryRow(`
        SELECT p.id, p.user_id, p.title, p.content, p.imagepath, 
               p.post_at, p.likes, p.dislikes, p.comments,
               u.username, u.profile_pic
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = ?
    `, id)

	var post utils.Post
	var postTime time.Time

	err := row.Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.ImagePath,
		&postTime,
		&post.Likes,
		&post.Dislikes,
		&post.Comments,
		&post.Username,
		&post.ProfilePic,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	post.PostTime = FormatTimeAgo(postTime)
	return &post, nil
}

func (ph *PostHandler) checkAuthStatus(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	_, err = utils.ValidateSession(utils.GlobalDB, cookie.Value)
	return err == nil
}
