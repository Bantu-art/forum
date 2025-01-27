package utils

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"golang.org/x/exp/rand"
)
var (
    ErrActiveSession = fmt.Errorf("user already has an active session")
    ErrNoSession    = fmt.Errorf("no active session found")
)

func GenerateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}



func HasActiveSession(db *sql.DB, userID string) (bool, error) {
    var exists bool
    err := db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM sessions 
            WHERE user_id = ? AND expires_at > ?
        )
    `, userID, time.Now()).Scan(&exists)
    
    if err != nil {
        return false, fmt.Errorf("error checking session: %v", err)
    }
    return exists, nil
}

func CreateSession(db *sql.DB, userID string) (string, error) {
    // Check for active session first
    hasSession, err := HasActiveSession(db, userID)
    if err != nil {
        return "", err
    }
    if hasSession {
        return "", ErrActiveSession
    }

    // Generate new session
    SessionToken := GenerateSessionToken()
    ExpiresAt := time.Now().Add(24 * time.Hour)

    // Create new session
    _, err = db.Exec(`
        INSERT INTO sessions(id, user_id, expires_at)
        VALUES (?, ?, ?)
    `, SessionToken, userID, ExpiresAt)
    if err != nil {
        return "", fmt.Errorf("failed to create session: %v", err)
    }

    log.Printf("Created new session for user %s", userID)
    return SessionToken, nil
}

func ValidateSession(db *sql.DB, sessionToken string) (string, error) {
	var userID string
	err := db.QueryRow(`
		SELECT user_id FROM sessions 
		WHERE id = ? AND expires_at > ?
	`, sessionToken, time.Now()).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("session expired or invalid")
		}
		return "", fmt.Errorf("error validating session: %v", err)
	}
	return userID, nil
}

func DeleteExpiredSessions(db *sql.DB) (int64, error) {
	result, err := db.Exec(`
		DELETE FROM sessions
		WHERE expires_at < ?
	`, time.Now())
	if err != nil {
		return 0, err
	}

	deletedSessions, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return deletedSessions, nil
}

func StartSessionsCLeanUp(ctx context.Context, db *sql.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rowsAffected, err := DeleteExpiredSessions(db)
				if err != nil {
					log.Printf("Failed to clean up expired sessions: %v", err)
				} else if rowsAffected > 0 {
					log.Printf("Cleaned up %d expired sessions", rowsAffected)
				}
			case <-ctx.Done():
				log.Println("Stopping session cleanup goroutine")
				return
			}
		}
	}()
}

func InitSessionManager(db *sql.DB) {
	ctx := context.Background()
	interval := 1 * time.Hour
	StartSessionsCLeanUp(ctx, db, interval)
}
