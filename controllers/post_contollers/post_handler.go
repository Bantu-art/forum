package postcontollers

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
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
        SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.imagepath, 
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

		// Convert postTime to local time before formatting
		post.PostTime = FormatTimeAgo(postTime.Local())
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (ph *PostHandler) checkAuthStatus(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	_, err = utils.ValidateSession(utils.GlobalDB, cookie.Value)
	return err == nil
}
