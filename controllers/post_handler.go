package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
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

		// Store userID in request context
		ctx := context.WithValue(r.Context(), "userID", userID)
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
	case "/react":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleReactions).ServeHTTP(w, r)
		} else {
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

// Update handler signatures to match http.HandlerFunc
func (ph *PostHandler) displayCreateForm(w http.ResponseWriter, r *http.Request) {
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
               u.username, u.profile_pic
        FROM posts p
        JOIN users u ON p.user_id = u.id
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
	currentTime := time.Now()
	result, err := stmt.Exec(userID, title, content, imagePath, currentTime)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()
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
    
    postIDStr := r.URL.Query().Get("id")
    
    if postIDStr == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }

    postID, err := strconv.ParseInt(postIDStr, 10, 64)
    if err != nil {
        log.Printf("Invalid post ID: %v", err)
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }

    // Get user ID if logged in
    var userID string
    if cookie, err := r.Cookie("session_token"); err == nil {
        userID, _ = utils.ValidateSession(utils.GlobalDB, cookie.Value)
    }

    post, err := ph.getPostByID(postID)
    if err != nil {
        log.Printf("Error fetching post: %v", err)
        http.Error(w, "Error fetching post", http.StatusInternalServerError)
        return
    }

    if post == nil {
        log.Printf("Post not found: %d", postID)
        http.Error(w, "Post not found", http.StatusNotFound)
        return
    }

    // Add user's reaction if logged in
    if userID != "" {
        var reaction int
        err := utils.GlobalDB.QueryRow(
            "SELECT like FROM reaction WHERE user_id = ? AND post_id = ?", 
            userID, postID,
        ).Scan(&reaction)
        if err != sql.ErrNoRows {
            post.UserReaction = &reaction
        }
    }


	tmpl, err := template.ParseFiles("templates/post.html")
    if err != nil {
        log.Printf("Template parsing error: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    if err := tmpl.Execute(w, post); err != nil {
        log.Printf("Template execution error: %v", err)
        // Don't write header here since Execute might have already written response
        log.Printf("Error rendering template: %v", err)
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

func (ph *PostHandler) handleReactions(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)
    if userID == "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
        return
    }

    var req struct {
        PostID int `json:"post_id"`
        Like   int `json:"like"` // 1 for like, 0 for dislike
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
        return
    }

    if req.Like != 0 && req.Like != 1 {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Invalid reaction type"})
        return
    }

    // Check if the user already has a reaction
    var existingLike int
    err := utils.GlobalDB.QueryRow("SELECT like FROM reaction WHERE user_id = ? AND post_id = ?", userID, req.PostID).Scan(&existingLike)
    if err != nil && err != sql.ErrNoRows {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
        return
    }

    if err == sql.ErrNoRows {
        // Insert new reaction
        _, err = utils.GlobalDB.Exec("INSERT INTO reaction (user_id, post_id, like) VALUES (?, ?, ?)", userID, req.PostID, req.Like)
        if err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
            return
        }
    } else {
        if existingLike == req.Like {
            // User is unliking or undisliking
            _, err = utils.GlobalDB.Exec("DELETE FROM reaction WHERE user_id = ? AND post_id = ?", userID, req.PostID)
            if err != nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
                return
            }
        } else {
            // Update existing reaction
            _, err = utils.GlobalDB.Exec("UPDATE reaction SET like = ? WHERE user_id = ? AND post_id = ?", req.Like, userID, req.PostID)
            if err != nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
                return
            }
        }
    }

    // Fetch updated like and dislike counts
    var likes, dislikes int
    err = utils.GlobalDB.QueryRow("SELECT likes, dislikes FROM posts WHERE id = ?", req.PostID).Scan(&likes, &dislikes)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
        return
    }

    // Return updated counts as JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]int{
        "likes":    likes,
        "dislikes": dislikes,
    })
}

