package controllers

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	// "net/url"
	"os"
	"strings"
	"time"

	"social-sync-backend/middleware"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// MastodonAppInfo stores the app registration details for each instance
type MastodonAppInfo struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
}

// Store app registrations temporarily (in production, use Redis or database)
var mastodonApps = make(map[string]*MastodonAppInfo)

// Store OAuth states temporarily (in production, use Redis or database)
var mastodonStates = make(map[string]string) // state -> instance_url|user_id

// Register app with Mastodon instance using JSON POST for better compatibility
func registerMastodonApp(instanceURL string) (*MastodonAppInfo, error) {
	if app, exists := mastodonApps[instanceURL]; exists {
		return app, nil
	}

	redirectURI := os.Getenv("MASTODON_REDIRECT_URL")
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/auth/mastodon/callback"
	}

	// Prepare JSON body
	bodyMap := map[string]string{
		"client_name":   "SocialSync",
		"redirect_uris": redirectURI,
		"scopes":        "read write",
		"website":       "https://yourdomain.com",
	}

	jsonBody, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal app registration JSON: %v", err)
	}

	req, err := http.NewRequest("POST", instanceURL+"/api/v1/apps", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create app registration request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register app: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("app registration failed: status %d, response: %s", resp.StatusCode, string(respBody))
	}

	var appInfo MastodonAppInfo
	if err := json.Unmarshal(respBody, &appInfo); err != nil {
		return nil, fmt.Errorf("failed to decode app registration response: %v", err)
	}

	appInfo.RedirectURI = redirectURI
	mastodonApps[instanceURL] = &appInfo
	return &appInfo, nil
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func normalizeInstanceURL(instanceURL string) string {
	instanceURL = strings.TrimSpace(instanceURL)
	if instanceURL == "" {
		return ""
	}
	if !strings.HasPrefix(instanceURL, "http://") && !strings.HasPrefix(instanceURL, "https://") {
		instanceURL = "https://" + instanceURL
	}
	return strings.TrimSuffix(instanceURL, "/")
}

func MastodonRedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appUserIDStr, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "Unauthorized: User not authenticated.", http.StatusUnauthorized)
			return
		}
		if _, err := uuid.Parse(appUserIDStr); err != nil {
			http.Error(w, "Internal server error: Invalid user ID format.", http.StatusInternalServerError)
			return
		}

		instance := r.URL.Query().Get("instance")
		if instance == "" {
			http.Error(w, "Missing instance parameter", http.StatusBadRequest)
			return
		}
		instanceURL := normalizeInstanceURL(instance)
		if instanceURL == "" {
			http.Error(w, "Invalid instance URL", http.StatusBadRequest)
			return
		}

		appInfo, err := registerMastodonApp(instanceURL)
		if err != nil {
			log.Printf("Failed to register Mastodon app: %v", err)
			http.Error(w, "Failed to register with Mastodon instance", http.StatusInternalServerError)
			return
		}

		state := generateState()
		mastodonStates[state] = instanceURL + "|" + appUserIDStr

		config := &oauth2.Config{
			ClientID:     appInfo.ClientID,
			ClientSecret: appInfo.ClientSecret,
			RedirectURL:  appInfo.RedirectURI,
			Scopes:       []string{"read", "write"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  instanceURL + "/oauth/authorize",
				TokenURL: instanceURL + "/oauth/token",
			},
		}

		authURL := config.AuthCodeURL(state)
		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	}
}

func MastodonCallbackHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state == "" {
			http.Error(w, "Missing state parameter", http.StatusBadRequest)
			return
		}
		stateData, exists := mastodonStates[state]
		if !exists {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}
		delete(mastodonStates, state)

		parts := strings.Split(stateData, "|")
		if len(parts) != 2 {
			http.Error(w, "Invalid state data", http.StatusBadRequest)
			return
		}
		instanceURL := parts[0]
		appUserIDStr := parts[1]

		if _, err := uuid.Parse(appUserIDStr); err != nil {
			http.Error(w, "Invalid user ID in state parameter", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code parameter", http.StatusBadRequest)
			return
		}

		appInfo, exists := mastodonApps[instanceURL]
		if !exists {
			http.Error(w, "App not registered for this instance", http.StatusInternalServerError)
			return
		}

		config := &oauth2.Config{
			ClientID:     appInfo.ClientID,
			ClientSecret: appInfo.ClientSecret,
			RedirectURL:  appInfo.RedirectURI,
			Scopes:       []string{"read", "write"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  instanceURL + "/oauth/authorize",
				TokenURL: instanceURL + "/oauth/token",
			},
		}

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		client := config.Client(context.Background(), token)
		userResp, err := client.Get(instanceURL + "/api/v1/accounts/verify_credentials")
		if err != nil {
			http.Error(w, "Failed to fetch user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer userResp.Body.Close()

		var userData struct {
			ID          string `json:"id"`
			Username    string `json:"username"`
			DisplayName string `json:"display_name"`
			Avatar      string `json:"avatar"`
			URL         string `json:"url"`
		}
		if err := json.NewDecoder(userResp.Body).Decode(&userData); err != nil {
			http.Error(w, "Failed to decode user data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var expiresAt *time.Time
		if token.Expiry != (time.Time{}) {
			expiresAt = &token.Expiry
		}

		socialID := fmt.Sprintf("%s:%s", instanceURL, userData.ID)
		profileName := userData.DisplayName
		if profileName == "" {
			profileName = userData.Username
		}
		profileName = fmt.Sprintf("%s (@%s)", profileName, userData.Username)

		_, err = db.Exec(`
			INSERT INTO social_accounts (
				user_id, platform, social_id, access_token, access_token_expires_at,
				refresh_token, profile_picture_url, profile_name, connected_at
			) VALUES (
				$1, 'mastodon', $2, $3, $4, $5, $6, $7, NOW()
			)
			ON CONFLICT (user_id, platform) DO UPDATE SET
				access_token = EXCLUDED.access_token,
				access_token_expires_at = EXCLUDED.access_token_expires_at,
				refresh_token = EXCLUDED.refresh_token,
				social_id = EXCLUDED.social_id,
				profile_picture_url = EXCLUDED.profile_picture_url,
				profile_name = EXCLUDED.profile_name,
				connected_at = NOW()
		`,
			appUserIDStr,
			socialID,
			token.AccessToken,
			expiresAt,
			token.RefreshToken,
			userData.Avatar,
			profileName,
		)
		if err != nil {
			http.Error(w, "Failed to save Mastodon account: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "http://localhost:3000/home/manage-accounts?connected=mastodon", http.StatusSeeOther)
	}
}
