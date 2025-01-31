package middleware

import (
    "context"
    "database/sql"
    "forum/utils"
    "log"
    "net/http"
    "time"
)

// SessionContextKey is used to store and retrieve the session from context
type SessionContextKey string

const (
    SessionKey   SessionContextKey = "session"
    AuthStateKey SessionContextKey = "authenticated"
)

// AuthMiddleware handles session validation and context injection
func AuthMiddleware(db *sql.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            ctx = context.WithValue(ctx, AuthStateKey, false)

            cookie, err := r.Cookie("session_token")

            session, err := GetSessionFromDB(db, cookie.Value)
            if err != nil {
                clearSessionCookie(w)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }

            if time.Now().After(session.ExpiresAt) {
                _, _ = db.Exec("DELETE FROM sessions WHERE id = ?", session.ID)
                clearSessionCookie(w)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }

            ctx = context.WithValue(ctx, SessionKey, session)
            ctx = context.WithValue(ctx, AuthStateKey, true)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func clearSessionCookie(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{
        Name:     "session_token",
        Value:    "",
        Path:     "/",
        Expires:  time.Now(),
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
}

// LoggingMiddleware logs request details and timing
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        lrw := &loggingResponseWriter{
            ResponseWriter: w,
            statusCode:     http.StatusOK,
        }

        next.ServeHTTP(lrw, r)

        isAuth := r.Context().Value(AuthStateKey).(bool)
        authStatus := "anonymous"
        if isAuth {
            authStatus = "authenticated"
        }

        log.Printf(
            "[%s] %s %s %s - Status: %d - Duration: %v",
            authStatus,
            r.Method,
            r.RequestURI,
            r.RemoteAddr,
            lrw.statusCode,
            time.Since(start),
        )
    })
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

// Helper functions to get session info from context

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