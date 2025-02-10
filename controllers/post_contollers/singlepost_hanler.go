package postcontollers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"forum/utils"
)

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
