package controllers

import (
	// "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"social-sync-backend/middleware"
	// "strings"
	// "time"
)

func ConnectInstagramHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Instagram connect handler started")

		userID, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			log.Println("Unauthorized:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var pageID, fbAccessToken string
		err = db.QueryRow(`
			SELECT social_id, access_token FROM social_accounts
			WHERE user_id = $1 AND platform = 'facebook'
		`, userID).Scan(&pageID, &fbAccessToken)
		if err != nil {
			log.Println("Facebook page not connected:", err)
			http.Error(w, "Facebook Page not connected", http.StatusBadRequest)
			return
		}

		// Step 1: Get IG Business ID
		graphURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=instagram_business_account&access_token=%s", pageID, fbAccessToken)
		resp, err := http.Get(graphURL)
		if err != nil {
			log.Println("Graph API error:", err)
			http.Error(w, "Facebook Graph API error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var igResp struct {
			InstagramBusinessAccount struct {
				ID string `json:"id"`
			} `json:"instagram_business_account"`
			Error struct {
				Message   string `json:"message"`
				Type      string `json:"type"`
				Code      int    `json:"code"`
				FBTraceID string `json:"fbtrace_id"`
			} `json:"error"`
		}
		var rawIGBody []byte
		rawIGBody, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read IG business ID response:", err)
			http.Error(w, "Failed to read IG business ID response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(rawIGBody, &igResp); err != nil {
			log.Println("Failed to parse IG business ID response:", err)
			log.Println("Raw IG business ID response:", string(rawIGBody))
			http.Error(w, "Failed to parse IG business ID response: "+err.Error()+"\nRaw: "+string(rawIGBody), http.StatusInternalServerError)
			return
		}
		if igResp.Error.Message != "" || igResp.InstagramBusinessAccount.ID == "" {
			log.Println("Instagram not linked to Facebook Page. Error:", igResp.Error.Message, igResp.Error.Type, igResp.Error.Code, igResp.Error.FBTraceID)
			log.Println("Raw IG business ID response:", string(rawIGBody))
			http.Error(w, "Instagram not linked to Facebook Page. Error: "+igResp.Error.Message+"\nRaw: "+string(rawIGBody), http.StatusBadRequest)
			return
		}
		igID := igResp.InstagramBusinessAccount.ID

		// Step 2: Fetch IG profile info
		profileURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=username,profile_picture_url&access_token=%s", igID, fbAccessToken)
		profileResp, err := http.Get(profileURL)
		if err != nil {
			http.Error(w, "Failed to fetch Instagram profile", http.StatusInternalServerError)
			return
		}
		defer profileResp.Body.Close()

		var profileData struct {
			Username          string `json:"username"`
			ProfilePictureURL string `json:"profile_picture_url"`
		}
		if err := json.NewDecoder(profileResp.Body).Decode(&profileData); err != nil {
			http.Error(w, "Failed to decode IG profile", http.StatusInternalServerError)
			return
		}

		// Step 3: Insert or update Instagram account
		_, err = db.Exec(`
			INSERT INTO social_accounts (
				user_id, platform, social_id, access_token,
				profile_name, profile_picture_url, connected_at
			) VALUES (
				$1, 'instagram', $2, $3, $4, $5, NOW()
			)
			ON CONFLICT (user_id, platform) DO UPDATE SET
				social_id = EXCLUDED.social_id,
				access_token = EXCLUDED.access_token,
				profile_name = EXCLUDED.profile_name,
				profile_picture_url = EXCLUDED.profile_picture_url,
				connected_at = NOW()
		`,
			userID,
			igID,
			fbAccessToken,
			profileData.Username,
			profileData.ProfilePictureURL,
		)
		if err != nil {
			http.Error(w, "Failed to save Instagram account", http.StatusInternalServerError)
			return
		}

		log.Println("Instagram connected successfully")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Instagram connected successfully",
		})
	}
}
