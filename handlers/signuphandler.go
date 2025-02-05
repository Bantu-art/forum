package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"forum/utils"

	_ "github.com/mattn/go-sqlite3"
)

type SignUpErrors struct {
	NameError     string
	EmailError    string
	UsernameError string
	PasswordError string
	GeneralError  string
}

type SignUpData struct {
	Errors   SignUpErrors
	Email    string
	UserName string
}

var GlobalDB *sql.DB

func InitDB(database *sql.DB) {
	GlobalDB = database
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/signup.html"))
		tmpl.Execute(w, &SignUpData{})
		return
	}

	if r.Method == "POST" {
		data := SignUpData{
			UserName: r.FormValue("username"),
			Email:    r.FormValue("email"),
		}
		errors := SignUpErrors{}
		hasError := false

		if !utils.ValidateEmail(data.Email) && data.Email != "" {
			errors.EmailError = "Invalid email format"
			hasError = true
		} else {
			errors.EmailError = "Email must be provided"
			hasError = true
		}

		if !utils.ValidateUsername(data.UserName) {
			errors.UsernameError = "Username must be between 3 and 30 characters, must"
			hasError = true
		}

		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		if !utils.ValidatePassword(password) {
			errors.PasswordError = "Password must be at least 8 characters, comprising of capital and small letters, numbers, and special characters"
			hasError = true
		} else if password != confirmPassword {
			errors.PasswordError = "Passwords do not match"
			hasError = true
		}

		if hasError {
			data.Errors = errors
			tmpl := template.Must(template.ParseFiles("templates/signup.html"))
			tmpl.Execute(w, data)
			return
		}

		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			errors.GeneralError = "Internal Server Error"
			data.Errors = errors
			tmpl := template.Must(template.ParseFiles("templates/signup.html"))
			tmpl.Execute(w, data)
			return
		}

		id := utils.GenerateId()

		_, err = GlobalDB.Exec(`
            INSERT INTO users (id, email, username, password)
            VALUES (?, ?, ?, ?)
        `, id, data.Email, data.UserName, hashedPassword)
		if err != nil {
			errors.GeneralError = "Username or email already exists"
			data.Errors = errors
			tmpl, err := template.ParseFiles("templates/signup.html")
			if err != nil {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				log.Printf("Error loading template: %v", err)
				return
			}
			tmpl.Execute(w, data)
			return
		}
	}

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}
