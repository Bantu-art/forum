package utils

import (
	"html/template"

	"net/http"
)

var ExecuteTemplate = func(w http.ResponseWriter, tmpl *template.Template, data interface{}) error {

	return tmpl.Execute(w, data)

}
