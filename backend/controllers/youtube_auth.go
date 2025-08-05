package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"social-sync-backend/middleware"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func getYouTubeOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("YOUTUBE_REDIRECT_URI"),
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube.upload",
			"https://www.googleapis.com/auth/youtube.readonly",
			"https://www.googleapis.com/auth/youtube.force-ssl",
		},
		Endpoint: google.Endpoint,
	}
}

func YouTubeRedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		config := getYouTubeOAuthConfig()
		state := fmt.Sprintf("%s:%d", userID, time.Now().UnixNano())
		url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

type YouTubeChannelInfo struct {
	Kind  string `json:"kind"`
	Items []struct {
		ID      string `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
		Statistics struct {
			ViewCount       string `json:"viewCount"`
			SubscriberCount string `json:"subscriberCount"`
			VideoCount      string `json:"videoCount"`
		} `json:"statistics"`
	} `json:"items"`
}

func YouTubeCallbackHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}

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
		userID := parts[0]
		if _, err := uuid.Parse(userID); err != nil {
			http.Error(w, "Invalid user ID in state parameter", http.StatusBadRequest)
			return
		}

		config := getYouTubeOAuthConfig()
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		client := config.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/youtube/v3/channels?part=snippet,statistics&mine=true")
		if err != nil {
			http.Error(w, "Failed to get channel info", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read channel info", http.StatusInternalServerError)
			return
		}

		var channelInfo YouTubeChannelInfo
		if err := json.Unmarshal(bodyBytes, &channelInfo); err != nil {
			http.Error(w, "Failed to decode channel info", http.StatusInternalServerError)
			return
		}

		if len(channelInfo.Items) == 0 {
			http.Error(w, "No YouTube channel found", http.StatusBadRequest)
			return
		}

		channel := channelInfo.Items[0]

		var existingAccountID string
		err = db.QueryRow(`
			SELECT id FROM social_accounts 
			WHERE user_id = $1 AND platform = 'youtube'
		`, userID).Scan(&existingAccountID)

		var expiresAt *time.Time
		if token.Expiry != (time.Time{}) {
			expiresAt = &token.Expiry
		}

		if err == sql.ErrNoRows {
			accountID := uuid.New()
			_, err = db.Exec(`
				INSERT INTO social_accounts (
					id, user_id, platform, social_id, access_token, 
					access_token_expires_at, refresh_token, profile_picture_url, 
					profile_name, connected_at, last_synced_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			`, accountID, userID, "youtube", channel.ID, token.AccessToken,
				expiresAt, token.RefreshToken, channel.Snippet.Thumbnails.Default.URL,
				channel.Snippet.Title, time.Now(), time.Now())

			if err != nil {
				http.Error(w, "Failed to save YouTube account", http.StatusInternalServerError)
				return
			}
		} else if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		} else {
			_, err = db.Exec(`
				UPDATE social_accounts 
				SET access_token = $1, access_token_expires_at = $2, refresh_token = $3,
					profile_picture_url = $4, profile_name = $5, last_synced_at = $6
				WHERE id = $7
			`, token.AccessToken, expiresAt, token.RefreshToken,
				channel.Snippet.Thumbnails.Default.URL, channel.Snippet.Title, time.Now(), existingAccountID)

			if err != nil {
				http.Error(w, "Failed to update YouTube account", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "http://localhost:3000/home/manage-accounts?connected=youtube", http.StatusSeeOther)
	}
}
