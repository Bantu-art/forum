package utils

import (
	"html/template"
	"net/http"
)

const (
	ErrMethodNotAllowed = "The requested method is not supported for this endpoint."
	ErrInternalServer   = "An unexpected error occurred. Please try again later."
	ErrUnauthorized     = "You must be logged in to perform this action."
	ErrForbidden        = "You don't have permission to perform this action."
	ErrInvalidForm      = "Please check your input and try again."
	ErrCategoryLoad     = "Unable to load categories. Please try again."
	ErrTemplateLoad     = "Unable to load page template."
	ErrPostCreate       = "Unable to create post. Please try again."
	ErrPostNotFound     = "The requested post could not be found."
	ErrCommentCreate    = "Unable to create comment. Please try again."
	ErrFileUpload       = "Error uploading file. Please try again."
	ErrPageNotFound     = "Page not found"
	ErrTemplateExec     = "We're experiencing technical difficulties. Please try again later."

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
