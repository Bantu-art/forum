package handlers

import (
	"net/http"
	"os"

	"forum/utils"
)

// Serve static files and avoid listing them to the web page
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	file := "." + r.URL.Path

	info, err := os.Stat(file)
	if err != nil {
		utils.RenderErrorPage(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	if info.IsDir() {
		utils.RenderErrorPage(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	http.ServeFile(w, r, file)
}
