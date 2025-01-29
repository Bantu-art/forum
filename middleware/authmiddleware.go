package middleware

import (
	"database/sql"
	"forum/utils"
)


// GetSessionFromDB retrieves a session from the database using the session token.
// It queries the database for a session with the given token and scans the result
// into a utils.Session struct. If the session is found, it returns the session and nil error.
// If an error occurs (e.g., session not found), it returns nil and the error.
func GetSessionFromDB(db *sql.DB, token string) (*utils.Session, error) {
	var session utils.Session
	err := db.QueryRow("SELECT id, user_id, expires_at FROM sessions WHERE token = ?", token).Scan(&session.ID, &session.UserID, &session.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
