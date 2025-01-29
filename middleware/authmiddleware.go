package middleware

import (
	"context"
	"database/sql"
	"forum/utils"
)

// SessionContextKey is used to store and retrieve the session from context
type SessionContextKey string

const (
	SessionKey   SessionContextKey = "session"
	AuthStateKey SessionContextKey = "authenticated"
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

// GetSession retrieves the session information from the provided context.
// It attempts to extract a value associated with the SessionKey from the context
// and asserts that this value is of type *utils.Session. If the value is not found
// or is nil, it returns nil and false. Otherwise, it returns the session and true,
// indicating that the session was successfully retrieved.
func GetSession(ctx context.Context) (*utils.Session, bool) {
	session, ok := ctx.Value(SessionKey).(*utils.Session)
	if !ok || session == nil {
		return nil, false
	}
	return session, true
}

// IsAuthenticated checks if a user is authenticated by looking for a boolean value
// associated with the AuthStateKey in the context. It returns true if the value is found
// and is true, indicating that the user is authenticated. If the value is not found or is false,
// it returns false, indicating that the user is not authenticated.
func IsAuthenticated(ctx context.Context) bool {
	auth, ok := ctx.Value(AuthStateKey).(bool)
	return ok && auth
}
