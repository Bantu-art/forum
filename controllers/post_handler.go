package controllers

import (
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

// Update your ServeHTTP method to include the new route
func (ph *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// if len(path) > 0 && path[len(path)-1] == '/' {
	// 	path = path[:len(path)-1]
	// }

	switch path {
	case "/create":
		if r.Method == http.MethodGet {
			ph.displayCreateForm(w)
		} else if r.Method == http.MethodPost {
			ph.handleCreatePost(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/":
		if r.Method == http.MethodGet {
			// Check if there's an ID parameter
			if r.URL.Query().Get("id") != "" {
				ph.handleSinglePost(w, r)
			} else {
				ph.handleGetPosts(w)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

// Add this new method to PostHandler
func (ph *PostHandler) displayCreateForm(w http.ResponseWriter) {
	tmpl, err := template.ParseFiles("templates/createpost.html")
	if err != nil {
		log.Printf("Error parsing create form template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error executing create form template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

func (ph *PostHandler) handleGetPosts(w http.ResponseWriter) {
	posts, err := ph.getAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, posts); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

func (ph *PostHandler) getAllPosts() ([]utils.Post, error) {
	rows, err := utils.GlobalDB.Query(`
        SELECT id, user_id, title, content,imagepath, post_at, likes, dislikes, comments 
        FROM posts 
        ORDER BY post_at DESC
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
		); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		post.PostTime = FormatTimeAgo(postTime)
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

// Update your handleCreatePost method
func (ph *PostHandler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	// Increase max memory for file uploads
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Error processing form", http.StatusBadRequest)
		return
	}

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")

	// Basic validation
	if title == "" {
		log.Printf("Title is empty")
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if content == "" {
		log.Printf("Content is empty")
		http.Error(w, "Content is required", http.StatusBadRequest)
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
	cookie, err := r.Cookie("session_token")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Validate session token and retrieve user ID
    userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
	currentTime := time.Now()
	result, err := stmt.Exec(userID, title, content, imagePath, currentTime)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()
	log.Printf("Successfully created post with ID: %d", postID)

	http.Redirect(w, r, "/post", http.StatusSeeOther)
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
        SELECT id, user_id, title, content, imagepath, post_at, likes, dislikes, comments 
        FROM posts 
        WHERE id = ?
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
