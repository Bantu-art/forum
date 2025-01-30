package controllers

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

func (ch *CategoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/categories":
		if r.Method == http.MethodGet {
			ch.handleGetCategories(w, r)
		} else if r.Method == http.MethodPost {
			ch.handleCreateCategory(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/categories/posts":
		if r.Method == http.MethodGet {
			categoryID := r.URL.Query().Get("id")
			if categoryID == "" {
				http.Error(w, "Category ID is required", http.StatusBadRequest)
				return
			}
			ch.handleGetPostsByCategory(w, categoryID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "/category":
		if r.Method == http.MethodGet {
			categoryName := r.URL.Query().Get("name")
			if categoryName == "" {
				http.Error(w, "Category name is required", http.StatusBadRequest)
				return
			}
			ch.handleGetPostsByCategoryName(w, categoryName)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

func (ch *CategoryHandler) handleGetCategories(w http.ResponseWriter, _ *http.Request) {
	categories, err := ch.getAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/categories.html")
	if err != nil {
		log.Printf("Error parsing categories template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, categories); err != nil {
		log.Printf("Error executing categories template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

func (ch *CategoryHandler) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error processing form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	stmt, err := utils.GlobalDB.Prepare("INSERT INTO categories (name) VALUES (?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, "Error creating category", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		http.Error(w, "Error creating category", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (ch *CategoryHandler) handleGetPostsByCategory(w http.ResponseWriter, categoryID string) {
	posts, err := ch.getPostsByCategory(categoryID)
	if err != nil {
		log.Printf("Error fetching posts for category %s: %v", categoryID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		log.Printf("Error parsing category posts template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, posts); err != nil {
		log.Printf("Error executing category posts template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

func (ch *CategoryHandler) handleGetPostsByCategoryName(w http.ResponseWriter, categoryName string) {
	posts, err := ch.getPostsByCategoryName(categoryName)
	if err != nil {
		log.Printf("Error fetching posts for category %s: %v", categoryName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		log.Printf("Error parsing category posts template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, posts); err != nil {
		log.Printf("Error executing category posts template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
	}
}

func (ch *CategoryHandler) getPostsByCategory(categoryID string) ([]utils.Post, error) {
	rows, err := utils.GlobalDB.Query(`
        SELECT p.id, p.title, p.content 
        FROM posts p
        JOIN post_categories pc ON p.id = pc.post_id
        WHERE pc.category_id = ?
    `, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []utils.Post
	for rows.Next() {
		var post utils.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (ch *CategoryHandler) getPostsByCategoryName(categoryName string) ([]utils.Post, error) {
	rows, err := utils.GlobalDB.Query(`
        SELECT p.id, p.title, p.content 
        FROM posts p
        JOIN post_categories pc ON p.id = pc.post_id
        JOIN categories c ON pc.category_id = c.id
        WHERE c.name = ?
    `, categoryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []utils.Post
	for rows.Next() {
		var post utils.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
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
