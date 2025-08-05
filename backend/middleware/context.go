package middleware

import (
    // "context"
    "net/http"
    "fmt"
    "log"
)

type contextKey string

const UserIDKey = contextKey("userID")

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(r *http.Request) (string, error) {
    if ctxUserID, ok := r.Context().Value(UserIDKey).(string); ok && ctxUserID != "" {
        log.Printf("DEBUG (GetUserID): Found userID from request context: %s", ctxUserID)
        return ctxUserID, nil
    }

    return "", fmt.Errorf("user ID not found in context; ensure authentication middleware is active")
}
