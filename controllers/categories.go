package controllers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"forum/utils"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

func (ch *CategoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/categories":
		if r.Method == http.MethodGet {
			ch.handleGetCategories(w, r)
		} else if r.Method == http.MethodPost {
			ch.handleCreateCategory(w, r)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	case "/category":
		if r.Method == http.MethodGet {
			categoryName := r.URL.Query().Get("name")
			if categoryName == "" {
				utils.RenderErrorPage(w, http.StatusBadRequest, utils.ErrInvalidForm)
				return
			}
			ch.handleGetPostsByCategoryName(w, r, categoryName)
		} else {
			utils.RenderErrorPage(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed)
		}
	default:
		utils.RenderErrorPage(w, http.StatusNotFound, utils.ErrNotFound)
	}
}

func (ch *CategoryHandler) checkAuthStatus(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	_, err = utils.ValidateSession(utils.GlobalDB, cookie.Value)
	return err == nil
}

func (ch *CategoryHandler) handleGetCategories(w http.ResponseWriter, _ *http.Request) {
	categories, err := ch.getAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		log.Printf("Error parsing categories template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
		return
	}

	if err := tmpl.Execute(w, categories); err != nil {
		log.Printf("Error executing categories template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
	}
}

func (ch *CategoryHandler) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		utils.RenderErrorPage(w, http.StatusBadRequest, utils.ErrInvalidForm)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		utils.RenderErrorPage(w, http.StatusBadRequest, utils.ErrInvalidForm)
		return
	}

	stmt, err := utils.GlobalDB.Prepare("INSERT INTO categories (name) VALUES (?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (ch *CategoryHandler) handleGetPostsByCategoryName(w http.ResponseWriter, r *http.Request, categoryName string) {
	posts, err := ch.getPostsByCategoryName(categoryName)
	if err != nil {
		log.Printf("Error fetching posts for category %s: %v", categoryName, err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInternalServer)
		return
	}

	isLoggedIn := ch.checkAuthStatus(r)
	var currentUserID string

	if cookie, err := r.Cookie("session_token"); err == nil {
		if userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value); err == nil {
			currentUserID = userID
		}
	}

	data := struct {
		IsLoggedIn    bool
		Posts         []utils.Post
		CurrentUserID string
	}{
		IsLoggedIn:    isLoggedIn,
		Posts:         posts,
		CurrentUserID: currentUserID,
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		log.Printf("Error parsing category posts template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing category posts template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
	}
}

func (ch *CategoryHandler) getPostsByCategoryName(categoryName string) ([]utils.Post, error) {
	rows, err := utils.GlobalDB.Query(`
        SELECT p.id, p.title, p.content, p.imagepath, p.post_at, u.username, u.profile_pic,
               (SELECT COUNT(*) FROM reaction WHERE post_id = p.id AND like = 1) AS Likes,
               (SELECT COUNT(*) FROM reaction WHERE post_id = p.id AND like = 0) AS Dislikes
        FROM posts p
        JOIN post_categories pc ON p.id = pc.post_id
        JOIN users u ON p.user_id = u.id
        JOIN categories c ON pc.category_id = c.id
        WHERE c.name = ?
    `, categoryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	postMap := make(map[int]utils.Post)
	for rows.Next() {
		var post utils.Post
		var postTime time.Time
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImagePath, &postTime, &post.Username, &post.ProfilePic, &post.Likes, &post.Dislikes); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
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

func (ch *CategoryHandler) getAllCategories() ([]utils.Category, error) {
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
