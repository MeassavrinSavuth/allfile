package controllers

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "golang.org/x/crypto/bcrypt"
    "social-sync-backend/lib"
    "social-sync-backend/models"
)

// EnableCORS (Note: This function seems to be misplaced. It's usually in your router/middleware setup, not in a controllers file as a standalone func. Assuming it's here for context)
func EnableCORS(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight
        if r.Method == http.MethodOptions {
            return
        }

        h.ServeHTTP(w, r)
    })
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    var user models.User // user.ID is now uuid.UUID
    err := lib.DB.QueryRow("SELECT id, password FROM users WHERE email = $1", req.Email).Scan(&user.ID, &user.Password)

    switch {
    case err == sql.ErrNoRows:
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    case err != nil:
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // If the user was created with Google and has no password set
    if user.Password == "" {
        http.Error(w, "This account uses Google login. Please sign in with Google.", http.StatusForbidden)
        return
    }

    // Check password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate tokens
    // user.ID is uuid.UUID, convert to string for token generation (JWT claims are usually strings)
    accessToken, err := lib.GenerateAccessToken(user.ID.String()) 
    if err != nil {
        http.Error(w, "Could not generate token", http.StatusInternalServerError)
        return
    }

    refreshToken, err := lib.GenerateRefreshToken(user.ID.String())
    if err != nil {
        http.Error(w, "Could not generate refresh token", http.StatusInternalServerError)
        return
    }

    // Return tokens
    json.NewEncoder(w).Encode(LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    })
}