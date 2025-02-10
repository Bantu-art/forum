package controllers

import (
	"html/template"
	"log"
	"net/http"

	"forum/utils"
)

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
