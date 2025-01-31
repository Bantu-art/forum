package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

func SignOutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Remove the session from the database
		_, err = db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		if err != nil {
			log.Printf("Error deleting session from database: %v", err)
		}

		// Clear the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now().Add(-1 * time.Hour),
		})

		// Redirect to the home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
