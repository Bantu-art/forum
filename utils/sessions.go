package utils

import (
	"database/sql"
	"encoding/base64"
	"time"

	"golang.org/x/exp/rand"
)

func GenerateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func CreateSession(db *sql.DB, userID string) (string, error) {
	SessionToken := GenerateSessionToken()
	ExpiresAt := time.Now().Add(24 * time.Hour)
	_, err := GlobalDB.Exec(`
	INSERT INTO sessions(id, user_id, expires_at)
	VALUES (?, ?, ?)
	`, SessionToken, userID, ExpiresAt)
	if err != nil {
		return "", err
	}

	return SessionToken, nil
}

func DeleteExpiredSessions(db *sql.DB) (int64, error) {
	result, err := db.Exec(
		`DELETE FROM sessions
		WHERE expires_at = ?
		`, time.Now())
	if err != nil {
		return 0, err
	}

	deletedSession, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return deletedSession, nil
}
