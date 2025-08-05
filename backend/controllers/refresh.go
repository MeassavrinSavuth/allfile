package controllers

import (
	"encoding/json"
	"net/http"
	"os"

	"social-sync-backend/lib"
)

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		http.Error(w, "Refresh token required", http.StatusBadRequest)
		return
	}

	claims, err := lib.VerifyToken(body.RefreshToken, os.Getenv("JWT_REFRESH_SECRET"))
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	userID := claims["user_id"].(string)
	accessToken, err := lib.GenerateAccessToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
}
