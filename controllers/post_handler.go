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
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/react":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleReactions).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/":
		if r.Method == http.MethodGet {
			if r.URL.Query().Get("id") != "" {
				ph.handleSinglePost(w, r)
			} else {
				ph.handleGetPosts(w, r)
			}
		}
	case "/comment":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleComment).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/commentreact":
		if r.Method == http.MethodPost {
			ph.authMiddleware(ph.handleCommentReactions).ServeHTTP(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	default:
		utils.RenderErrorPage(w, http.StatusNotFound, utils.ErrPageNotFound)
	}
}

func (ph *PostHandler) displayCreateForm(w http.ResponseWriter, r *http.Request) {
	categories, err := ph.getAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrCategoryLoad)
		return
	}

	tmpl, err := template.ParseFiles("templates/createpost.html")
	if err != nil {
		log.Printf("Error parsing create form template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateLoad)
		return
	}

	data := struct {
		Categories    []utils.Category
		CurrentUserID string
		IsLoggedIn    bool
	}{
		Categories: categories,
		IsLoggedIn: ph.checkAuthStatus(r),
	}
	if cookie, err := r.Cookie("session_token"); err == nil {
		if userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value); err == nil {
			data.CurrentUserID = userID
		}
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing create form template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
		return
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
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
		return
	}

	pageData := utils.PageData{
		IsLoggedIn: ph.checkAuthStatus(r),
		Posts:      posts,
	}

	if cookie, err := r.Cookie("session_token"); err == nil {
		if userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value); err == nil {
			pageData.CurrentUserID = userID
		}
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateLoad)
		return
	}

	if err := tmpl.Execute(w, pageData); err != nil {
		log.Printf("Error executing template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
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

	postMap := make(map[int]utils.Post)
	for rows.Next() {
		var post utils.Post
		var postTime time.Time

		var categoryID sql.NullInt64
		var categoryName sql.NullString
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
			&categoryID,
			&categoryName,
		); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}

		if categoryID.Valid {
			categoryIDInt := int(categoryID.Int64)
			post.CategoryID = &categoryIDInt
		} else {
			post.CategoryID = nil
		}

		if categoryName.Valid {
			post.CategoryName = &categoryName.String
		} else {
			post.CategoryName = nil
		}

		post.PostTime = FormatTimeAgo(postTime.Local())

		postMap[post.ID] = post
	}

	var posts []utils.Post
	for _, post := range postMap {
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (ph *PostHandler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		utils.RenderErrorPage(w, http.StatusUnauthorized, utils.ErrUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing form: %v", err)
		utils.RenderErrorPage(w, http.StatusBadRequest, utils.ErrInvalidForm)
		return
	}

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.Form["categories[]"]

	if title == "" || content == "" || len(categories) == 0 {
		log.Printf("Title, content, and category are required")
		utils.RenderErrorPage(w, http.StatusBadRequest, utils.ErrInvalidForm)
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
			utils.RenderErrorPage(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Prepare the insert statement with image support
	stmt, err := utils.GlobalDB.Prepare(`
        INSERT INTO posts (user_id, title, content, imagepath, post_at, likes, dislikes, comments) 
        VALUES (?, ?, ?, ?, ?, 0, 0, 0)
    `)
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}
	defer stmt.Close()

	// Execute the insert
	currentTime := time.Now()
	result, err := stmt.Exec(userID, title, content, imagePath, currentTime)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}

	postID, _ := result.LastInsertId()

	for _, categoryName := range categories {
		categoryID, err := getCategoryIDByName(categoryName)
		if err != nil {
			log.Printf("Error getting category ID for %s: %v", categoryName, err)
			continue
		}
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

func getCategoryIDByName(name string) (int, error) {
	var id int
	err := utils.GlobalDB.QueryRow("SELECT id FROM categories WHERE name = ?", name).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
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
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrPostNotFound)
		return
	}

	post, comments, err := ph.getPostByID(postID)
	if err != nil {
		log.Printf("Error fetching post: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrPostNotFound)
		return
	}

	if post == nil {
		log.Printf("Post not found: %d", postID)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrPostNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
		return
	}

	data := struct {
		Post          *utils.Post
		Comments      []utils.Comment
		CurrentUserID string
		IsLoggedIn    bool
	}{
		Post:       post,
		Comments:   comments,
		IsLoggedIn: ph.checkAuthStatus(r),
	}
	if cookie, err := r.Cookie("session_token"); err == nil {
		if userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value); err == nil {
			data.CurrentUserID = userID
		}
	}

	if err := tmpl.Execute(w, data); err != nil {
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
	}
}

// Add this helper method to fetch a single post
func (ph *PostHandler) getPostByID(id int64) (*utils.Post, []utils.Comment, error) {
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
		return nil, nil, err
	}
	if err != nil {
		return nil, nil, err
	}

	post.PostTime = FormatTimeAgo(postTime.Local())
	// Get comments
	rows, err := utils.GlobalDB.Query(`
	  SELECT c.id, c.content, c.comment_at, u.username, u.profile_pic 
	  FROM comments c
	  JOIN users u ON c.user_id = u.id
	  WHERE c.post_id = ?
	  ORDER BY c.comment_at DESC`, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var comments []utils.Comment
	for rows.Next() {
		var c utils.Comment
		var t time.Time
		err := rows.Scan(&c.ID, &c.Content, &t, &c.Username, &c.ProfilePic)
		if err != nil {
			continue
		}
		c.CommentTime = t.Local()
		comments = append(comments, c)
	}

	return &post, comments, nil
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

func (ph *PostHandler) handleComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	// Insert comment
	_, err = utils.GlobalDB.Exec(`
        INSERT INTO comments (post_id, user_id, content) 
        VALUES (?, ?, ?)`,
		postID, userID, content,
	)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update post comment count
	_, err = utils.GlobalDB.Exec(`
        UPDATE posts SET comments = comments + 1 
        WHERE id = ?`, postID)
	if err != nil {
		log.Printf("Error updating comment count: %v", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/?id=%d", postID), http.StatusSeeOther)
}

// handleCommentReactions processes a user's like/dislike on a comment.
func (ph *PostHandler) handleCommentReactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		CommentID int `json:"comment_id"`
		Like      int `json:"like"` // 1 for like, 0 for dislike
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

	// Check if the user already reacted to this comment by querying comment_reaction
	var existingIsLike int
	err := utils.GlobalDB.QueryRow("SELECT is_like FROM comment_reaction WHERE user_id = ? AND comment_id = ?", userID, req.CommentID).Scan(&existingIsLike)
	if err != nil && err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error (select)"})
		return
	}

	if err == sql.ErrNoRows {
		// No reaction exists—insert a new reaction.
		_, err = utils.GlobalDB.Exec("INSERT INTO comment_reaction (user_id, comment_id, is_like) VALUES (?, ?, ?)", userID, req.CommentID, req.Like)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error (insert)"})
			return
		}
	} else {
		if existingIsLike == req.Like {
			// Same reaction exists; remove it.
			_, err = utils.GlobalDB.Exec("DELETE FROM comment_reaction WHERE user_id = ? AND comment_id = ?", userID, req.CommentID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error (delete)"})
				return
			}
		} else {
			// Reaction exists but is different; update it.
			_, err = utils.GlobalDB.Exec("UPDATE comment_reaction SET is_like = ? WHERE user_id = ? AND comment_id = ?", req.Like, userID, req.CommentID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error (update)"})
				return
			}
		}
	}

	// Fetch updated like and dislike counts for the comment.
	var likes, dislikes int
	err = utils.GlobalDB.QueryRow("SELECT likes, dislikes FROM comments WHERE id = ?", req.CommentID).Scan(&likes, &dislikes)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error (display)"})
		return
	}

	// Return the updated counts.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}
