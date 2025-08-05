package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"social-sync-backend/middleware"
)

func getFacebookOAuthConfig() *oauth2.Config {
	redirectURL := os.Getenv("FACEBOOK_REDIRECT_URL")
	if redirectURL == "" {
		log.Fatal("FACEBOOK_REDIRECT_URL is empty!")
	}
	log.Printf("DEBUG: Using FACEBOOK_REDIRECT_URL: %s", redirectURL)

	return &oauth2.Config{
		ClientID:     os.Getenv("FACEBOOK_APP_ID"),
		ClientSecret: os.Getenv("FACEBOOK_APP_SECRET"),
		RedirectURL:  redirectURL,
		Scopes: []string{
			"email", "public_profile",
			"pages_show_list", "business_management",
			"pages_manage_posts", "instagram_basic",
			"instagram_content_publish", "pages_read_engagement",
			"read_insights", "instagram_manage_insights",
		},
		Endpoint: facebook.Endpoint,
	}
}

func FacebookRedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := getFacebookOAuthConfig()
		appUserIDStr, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "Unauthorized: User not authenticated.", http.StatusUnauthorized)
			return
		}

		if _, err := uuid.Parse(appUserIDStr); err != nil {
			http.Error(w, "Internal server error: Invalid user ID format.", http.StatusInternalServerError)
			return
		}

		state := fmt.Sprintf("%s:%d", appUserIDStr, time.Now().UnixNano())
		authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

func FacebookCallbackHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state == "" {
			http.Error(w, "Missing state parameter", http.StatusBadRequest)
			return
		}
		parts := strings.Split(state, ":")
		if len(parts) < 1 {
			http.Error(w, "Invalid state parameter format", http.StatusBadRequest)
			return
		}
		appUserIDStr := parts[0]
		if _, err := uuid.Parse(appUserIDStr); err != nil {
			http.Error(w, "Invalid user ID in state parameter", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		config := getFacebookOAuthConfig()
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		client := config.Client(context.Background(), token)

		pagesResp, err := client.Get("https://graph.facebook.com/v18.0/me/accounts")
		if err != nil {
			http.Error(w, "Failed to fetch pages: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer pagesResp.Body.Close()

		var pageData struct {
			Data []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				AccessToken string `json:"access_token"`
			} `json:"data"`
		}
		if err := json.NewDecoder(pagesResp.Body).Decode(&pageData); err != nil {
			http.Error(w, "Failed to decode page data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(pageData.Data) == 0 {
			http.Error(w, "No Facebook Pages found", http.StatusBadRequest)
			return
		}

		page := pageData.Data[0] // TODO: Let user pick later
		pageID := page.ID
		pageAccessToken := page.AccessToken
		pageName := page.Name
		pictureURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/picture?type=large", pageID)

		_, err = db.Exec(`
			INSERT INTO social_accounts (
				user_id, platform, social_id, access_token,
				profile_picture_url, profile_name, connected_at
			) VALUES (
				$1, 'facebook', $2, $3, $4, $5, NOW()
			)
			ON CONFLICT (user_id, platform) DO UPDATE SET
				access_token = EXCLUDED.access_token,
				social_id = EXCLUDED.social_id,
				profile_picture_url = EXCLUDED.profile_picture_url,
				profile_name = EXCLUDED.profile_name,
				connected_at = NOW()
		`,
			appUserIDStr,
			pageID,
			pageAccessToken,
			pictureURL,
			pageName,
		)
		if err != nil {
			http.Error(w, "Failed to save Facebook Page account: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "http://localhost:3000/home/manage-accounts?connected=facebook", http.StatusSeeOther)
	}
}

