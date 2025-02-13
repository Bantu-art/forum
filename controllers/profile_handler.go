package controllers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"forum/utils"
)

type ProfileHandler struct {
	imageHandler *ImageHandler
}

type ProfileData struct {
	Username     string
	Email        string
	ProfilePic   sql.NullString
	IsLoggedIn   bool
	IsOwnProfile bool
	UserID       string
}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{
		imageHandler: NewImageHandler(),
	}
}

func (ph *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract userID from URL path
	urlPath := r.URL.Path
	targetUserID := strings.TrimPrefix(urlPath, "/profile/")

	// Check if viewing own profile
	var currentUserID string
	isLoggedIn := false

	// Check if user is logged in
	if cookie, err := r.Cookie("session_token"); err == nil {
		if userID, err := utils.ValidateSession(utils.GlobalDB, cookie.Value); err == nil {
			currentUserID = userID
			isLoggedIn = true
		}
	}

	// Handle profile updates only for own profile
	if r.Method == "POST" && targetUserID == currentUserID {
		ph.handleProfileUpdate(w, r, currentUserID)
		return
	}

	// Display profile
	ph.displayUserProfile(w, targetUserID, currentUserID, isLoggedIn)
}

func (ph *ProfileHandler) displayUserProfile(w http.ResponseWriter, targetUserID string, currentUserID string, isLoggedIn bool) {
	var profile ProfileData
	err := utils.GlobalDB.QueryRow(`
        SELECT id, username, email, COALESCE(profile_pic, '') as profile_pic 
        FROM users 
        WHERE id = ?
    `, targetUserID).Scan(&profile.UserID, &profile.Username, &profile.Email, &profile.ProfilePic)
	if err != nil {
		if err == sql.ErrNoRows {
utils.RenderErrorPage(w, http.StatusNotFound, utils.ErrNotFound)	
		return
		}
		log.Printf("Error fetching profile: %v", err)
		utils.RenderErrorPage(w, http.StatusNotFound, utils.ErrNotFound)	
		return
	}

	profile.IsLoggedIn = isLoggedIn
	profile.IsOwnProfile = targetUserID == currentUserID

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		utils.RenderErrorPage(w, http.StatusInternalServerError, utils.ErrInternalServer)	
		return
	}

	tmpl.Execute(w, profile)
}

func (ph *ProfileHandler) handleProfileUpdate(w http.ResponseWriter, r *http.Request, userID string) {
	// Set max upload size - 5MB
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error processing form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("profile_pic")
if err != nil {
    log.Printf("Error getting profile pic: %v", err)
    utils.RenderErrorPage(w, http.StatusBadRequest, utils.ErrFileUpload)
    return
}
defer file.Close()

	// Validate file type
	if !isValidImageType(header.Header.Get("Content-Type")) {
		log.Printf("Invalid file type: %s", header.Header.Get("Content-Type"))
		http.Error(w, "Invalid file type. Please upload an image.", http.StatusBadRequest)
		return
	}

	// Get old profile pic path
	var oldImagePath sql.NullString
	err = utils.GlobalDB.QueryRow("SELECT profile_pic FROM users WHERE id = ?", userID).Scan(&oldImagePath)
	if err != nil {
		log.Printf("Error fetching old profile pic: %v", err)
	}

	// Process new image
	imagePath, err := ph.imageHandler.ProcessImage(file, header)
if err != nil {
    log.Printf("Error processing image: %v", err)
    utils.RenderErrorPage(w, http.StatusBadRequest, err.Error())
    return
}

	log.Printf("New image path: %s", imagePath)

	// Update database with new image path
	result, err := utils.GlobalDB.Exec(`
        UPDATE users 
        SET profile_pic = ? 
        WHERE id = ?
    `, imagePath, userID)
	if err != nil {
		log.Printf("Error updating profile pic in database: %v", err)
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		os.Remove(imagePath) // Clean up new image if database update fails
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Database update affected %d rows", rowsAffected)

	// Delete old profile pic if it exists
	if oldImagePath.Valid && oldImagePath.String != "" {
		oldFilePath := strings.TrimPrefix(oldImagePath.String, "/static")
		oldFilePath = filepath.Join("static", oldFilePath)
		if err := os.Remove(oldFilePath); err != nil {
			log.Printf("Error deleting old profile pic: %v", err)
		}
	}
	// fmt.Println("userID: %s", userID)

	// Redirect back to profile page with userID
	http.Redirect(w, r, "/profile/"+userID, http.StatusSeeOther)
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	return validTypes[contentType]
}
