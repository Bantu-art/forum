package utils

import (
	"html/template"
	"net/http"
)

func RenderErrorPage(w http.ResponseWriter, code int, message string) {
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := ErrorPageData{
		Code:    code,
		Message: message,
	}
	tmpl.Execute(w, data)
}
