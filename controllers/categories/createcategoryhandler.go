package categories

import (
	"log"
	"net/http"

	"forum/utils"
)

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
