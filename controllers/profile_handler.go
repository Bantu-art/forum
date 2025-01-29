package controllers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"forum/utils"
)

type ProfileHandler struct {
	imageHandler *ImageHandler
}

type ProfileData struct {
	Username   string
	Email      string
	ProfilePic sql.NullString
	IsLoggedIn bool
}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{
		imageHandler: NewImageHandler(),
	}
}

func (ph *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check session
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		ph.handleProfileUpdate(w, r, userID)
		return
	}

	ph.displayProfile(w, r, userID)
}

func (ph *ProfileHandler) displayProfile(w http.ResponseWriter, r *http.Request, userID string) {
	var profile ProfileData
    err := utils.GlobalDB.QueryRow(`
        SELECT username, email, COALESCE(profile_pic, '') as profile_pic 
        FROM users 
        WHERE id = ?
    `, userID).Scan(&profile.Username, &profile.Email, &profile.ProfilePic)

    if err != nil {
        log.Printf("Error fetching profile: %v", err)
        http.Error(w, "Error loading profile", http.StatusInternalServerError)
        return
    }

	profile.IsLoggedIn = true

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, profile)
}

func (ph *ProfileHandler) handleProfileUpdate(w http.ResponseWriter, r *http.Request, userID string) {
	file, header, err := r.FormFile("profile_pic")
	if err != nil {
		log.Printf("Error getting profile pic: %v", err)
		http.Error(w, "Error uploading image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imagePath, err := ph.imageHandler.ProcessImage(file, header)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		http.Error(w, "Error processing image", http.StatusInternalServerError)
		return
	}

	_, err = utils.GlobalDB.Exec(`
        UPDATE users 
        SET profile_pic = ? 
        WHERE id = ?
    `, imagePath, userID)
	if err != nil {
		log.Printf("Error updating profile pic: %v", err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
