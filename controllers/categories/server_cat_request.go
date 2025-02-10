package categories

import "net/http"

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
	case "/category":
		if r.Method == http.MethodGet {
			categoryName := r.URL.Query().Get("name")
			if categoryName == "" {
				http.Error(w, "Category name is required", http.StatusBadRequest)
				return
			}
			ch.handleGetPostsByCategoryName(w, r, categoryName)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}
