package categories

import (
	"html/template"
	"log"
	"net/http"

	"forum/utils"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

func (ch *CategoryHandler) handleGetPostsByCategoryName(w http.ResponseWriter, r *http.Request, categoryName string) {
	posts, err := ch.getPostsByCategoryName(categoryName)
	if err != nil {
		log.Printf("Error fetching posts for category %s: %v", categoryName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	isLoggedIn := ch.checkAuthStatus(r)

	data := utils.PageData{
		IsLoggedIn: isLoggedIn,
		Posts:      posts,
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		log.Printf("Error parsing category posts template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing category posts template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

func (ch *CategoryHandler) getPostsByCategoryName(categoryName string) ([]utils.Post, error) {
	rows, err := utils.GlobalDB.Query(`
        SELECT p.id, p.title, p.content, p.imagepath, u.username, u.profile_pic
        FROM posts p
        JOIN post_categories pc ON p.id = pc.post_id
		JOIN users u ON p.user_id = u.id
        JOIN categories c ON pc.category_id = c.name
        WHERE c.name = ?
    `, categoryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []utils.Post
	for rows.Next() {
		var post utils.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImagePath, &post.Username, &post.ProfilePic); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

