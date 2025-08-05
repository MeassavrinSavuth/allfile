package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"social-sync-backend/middleware"
	"github.com/gorilla/mux"
)

// GetSocialAccountsHandler fetches all social accounts linked to the authenticated user.
func GetSocialAccountsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		appUserIDVal := ctx.Value(middleware.UserIDKey)
		appUserID, ok := appUserIDVal.(string)
		if !ok || appUserID == "" {
			log.Println("ERROR: Unauthorized access to GetSocialAccountsHandler: User ID not found in context.")
			http.Error(w, "Unauthorized: User not authenticated.", http.StatusUnauthorized)
			return
		}

		rows, err := db.QueryContext(ctx, `
			SELECT platform, profile_picture_url, profile_name, social_id
			FROM social_accounts
			WHERE user_id = $1
		`, appUserID)
		if err != nil {
			log.Printf("ERROR: Failed to fetch social accounts for user %s: %v", appUserID, err)
			http.Error(w, "Internal server error: Could not fetch social accounts.", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type SocialAccountResponse struct {
			Platform          string  `json:"platform"`
			SocialID          string  `json:"socialId"`
			ProfilePictureURL *string `json:"profilePictureUrl"`
			ProfileName       *string `json:"profileName"`
		}
		var accounts []SocialAccountResponse

		for rows.Next() {
			var acc SocialAccountResponse
			if err := rows.Scan(&acc.Platform, &acc.ProfilePictureURL, &acc.ProfileName, &acc.SocialID); err != nil {
				log.Printf("ERROR: Error scanning social account row for user %s: %v", appUserID, err)
				http.Error(w, "Internal server error: Error scanning data.", http.StatusInternalServerError)
				return
			}
			accounts = append(accounts, acc)
		}

		if err := rows.Err(); err != nil {
			log.Printf("ERROR: Error after iterating through social account rows for user %s: %v", appUserID, err)
			http.Error(w, "Internal server error: Database iteration error.", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(accounts); err != nil {
			log.Printf("ERROR: Failed to encode social accounts to JSON for user %s: %v", appUserID, err)
			http.Error(w, "Internal server error: Could not encode response.", http.StatusInternalServerError)
			return
		}

		log.Printf("INFO: Successfully fetched %d social accounts for user %s.", len(accounts), appUserID)
	}
}

// DisconnectSocialAccountHandler unlinks a social media account for the authenticated user.
func DisconnectSocialAccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		appUserIDVal := ctx.Value(middleware.UserIDKey)
		appUserID, ok := appUserIDVal.(string)
		if !ok || appUserID == "" {
			log.Println("ERROR: Unauthorized access to DisconnectSocialAccountHandler: User ID not found in context.")
			http.Error(w, "Unauthorized: User not authenticated.", http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(r)
		platform := vars["platform"]
		if platform == "" {
			log.Println("ERROR: Platform name missing in request URL.")
			http.Error(w, "Bad request: Missing platform.", http.StatusBadRequest)
			return
		}

		// Normalize platform name for consistent matching
		platform = strings.ToLower(platform)
		if platform == "twitter (x)" {
			platform = "twitter"
		}

		log.Printf("DEBUG: Disconnecting platform '%s' for user %s", platform, appUserID)

		result, err := db.ExecContext(ctx, `
			DELETE FROM social_accounts
			WHERE user_id = $1 AND LOWER(platform) = $2
		`, appUserID, platform)

		if err != nil {
			log.Printf("ERROR: Failed to disconnect %s for user %s: %v", platform, appUserID, err)
			http.Error(w, "Internal server error: Failed to disconnect.", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("ERROR: RowsAffected error: %v", err)
			http.Error(w, "Internal server error.", http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			log.Printf("INFO: No %s account found for user %s to disconnect.", platform, appUserID)
			http.Error(w, "No such account connected.", http.StatusNotFound)
			return
		}

		log.Printf("INFO: Successfully disconnected %s for user %s.", platform, appUserID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Disconnected successfully",
		})
	}
}
