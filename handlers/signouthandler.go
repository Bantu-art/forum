package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"forum/utils"
)

func SignOutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		_, err = db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		if err != nil {
			log.Printf("Error deleting session from database: %v", err)
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now().Add(-1 * time.Hour),
		})

		if err != nil {
			utils.RenderErrorPage(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
