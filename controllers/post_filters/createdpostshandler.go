package postfilters

import (
	"log"
	"net/http"
	"time"

	postcontollers "forum/controllers/post_contollers"
	"forum/utils"
)

func CreatedPosts(w http.ResponseWriter, r *http.Request) {
	// Set content type header at the start
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Check session
	userID, err := validateUserSession(w, r)
	if err != nil {
		return
	}

	// Fetch posts
	posts, err := fetchUserPostsForPosts(userID)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	// Render template
	if err := renderCreatedTemplateForPosts(w, posts, userID); err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}
}

func validateUserSession(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return "", err
	}

	userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return "", err
	}

	return userID, nil
}

func fetchUserPostsForPosts(userID string) ([]utils.Post, error) {
	rows, err := utils.GlobalDB.Query(`
        SELECT DISTINCT
            p.id AS post_id,
            p.user_id,
            p.title,
            p.content,
            p.imagepath,
            p.post_at,
            p.likes,
            p.dislikes,
            p.comments,
            u.username,
            u.profile_pic,
            c.id AS category_id,
            c.name AS category_name
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN post_categories pc ON p.id = pc.post_id
        LEFT JOIN categories c ON pc.category_id = c.id
        WHERE p.user_id = ?
        ORDER BY p.post_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var post_time time.Time
	var posts []utils.Post
	for rows.Next() {
		var post utils.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&post_time,
			&post.Likes,
			&post.Dislikes,
			&post.Comments,
			&post.Username,
			&post.ProfilePic,
			&post.CategoryID,
			&post.CategoryName,
		)
		if err != nil {
			return nil, err
		}
		post.PostTime = postcontollers.FormatTimeAgo(post_time.Local())
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
