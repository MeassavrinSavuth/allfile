package controllers

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"time"

	"social-sync-backend/lib"
)

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message}) // CHANGED "error" -> "message"
}
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	var userID string
	err := lib.DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	var storedToken string
	var expires time.Time
	err = lib.DB.QueryRow("SELECT token, expires_at FROM email_verifications WHERE user_id = $1", userID).Scan(&storedToken, &expires)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Verification token not found")
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.Token), []byte(storedToken)) != 1 || time.Now().After(expires) {
		writeJSONError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	_, err = lib.DB.Exec("UPDATE users SET is_verified = TRUE WHERE id = $1", userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	_, _ = lib.DB.Exec("DELETE FROM email_verifications WHERE user_id = $1", userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Email verified successfully"})
}
