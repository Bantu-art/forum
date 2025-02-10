package postcontollers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"forum/utils"
)

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
	fmt.Println(title)
	content := r.FormValue("content")
	fmt.Println(content)
	categories := r.Form["categories[]"]

	if title == "" || content == "" || len(categories) == 0 {
		log.Printf("Title, content, and category are required")
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInvalidForm)
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
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
		return
	}
	defer stmt.Close()

	// Execute the insert
	currentTime := time.Now()
	result, err := stmt.Exec(userID, title, content, imagePath, currentTime)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrTemplateExec)
		return
	}

	postID, _ := result.LastInsertId()

	for _, categoryID := range categories {
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

func AFormatTimeAgo(t time.Time) string {
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
