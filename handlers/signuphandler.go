package handlers

import (
	"database/sql"
	"html/template"
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

		// Validate email
		if !utils.ValidateEmail(data.Email) {
			errors.EmailError = "Invalid email format"
			hasError = true
		}

		// Validate username
		if !utils.ValidateUsername(data.UserName) {
			errors.UsernameError = "Username must be between 3 and 30 characters"
			hasError = true
		}

		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		if !utils.ValidatePassword(password) {
			errors.PasswordError = "Password must be at least 8 characters"
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

		// Hash password and create user
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
			tmpl := template.Must(template.ParseFiles("templates/signup.html"))
			tmpl.Execute(w, data)
			return
		}
	}

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}
