package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"social-sync-backend/middleware"
)

type TwitterPostRequest struct {
	Message string `json:"message"`
}

type TwitterPostResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

type TwitterErrorResponse struct {
	Errors []struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"errors"`
}

func PostToTwitterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		var req TwitterPostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate message
		message := strings.TrimSpace(req.Message)
		if message == "" {
			http.Error(w, "Message cannot be empty", http.StatusBadRequest)
			return
		}

		// Check Twitter character limit (280 characters)
		if len(message) > 280 {
			http.Error(w, "Message exceeds Twitter's 280 character limit", http.StatusBadRequest)
			return
		}

		// Get Twitter account details
		var accessToken string
		var tokenExpiry *time.Time
		var refreshToken *string

		err = db.QueryRow(`
			SELECT access_token, access_token_expires_at, refresh_token
			FROM social_accounts
			WHERE user_id = $1 AND platform = 'twitter'
		`, userID).Scan(&accessToken, &tokenExpiry, &refreshToken)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Twitter account not connected", http.StatusBadRequest)
				return
			}
			http.Error(w, "Failed to retrieve Twitter account", http.StatusInternalServerError)
			return
		}

		// Debug: Print the access token (first 10 chars for privacy)
		fmt.Printf("DEBUG: Twitter access token (first 10 chars): %s\n", accessToken[:10])
		fmt.Printf("DEBUG: Twitter message content: %q\n", message)

		// Check if token is expired (if expiry is set)
		if tokenExpiry != nil && time.Now().After(*tokenExpiry) {
			// In a production app, you'd refresh the token here
			// For now, we'll return an error
			http.Error(w, "Twitter access token has expired. Please reconnect your account.", http.StatusUnauthorized)
			return
		}

		// Prepare tweet payload
		tweetPayload := map[string]interface{}{
			"text": message,
		}

		payloadBytes, err := json.Marshal(tweetPayload)
		if err != nil {
			http.Error(w, "Failed to prepare tweet payload", http.StatusInternalServerError)
			return
		}

		// Debug: Print the payload being sent
		fmt.Printf("DEBUG: Twitter payload: %s\n", string(payloadBytes))

		// Create HTTP request to Twitter API
		tweetURL := "https://api.twitter.com/2/tweets"
		req_twitter, err := http.NewRequest("POST", tweetURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		// Set headers - Use OAuth 2.0 Bearer token
		req_twitter.Header.Set("Content-Type", "application/json")
		req_twitter.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req_twitter.Header.Set("User-Agent", "SocialSync/1.0")
		req_twitter.Header.Set("Accept", "application/json")

		// Debug: Print request details
		fmt.Printf("DEBUG: Twitter request URL: %s\n", tweetURL)
		fmt.Printf("DEBUG: Twitter request headers: %v\n", req_twitter.Header)

		// Make the request
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Simple retry mechanism
		var resp *http.Response
		for attempt := 1; attempt <= 2; attempt++ {
			resp, err = client.Do(req_twitter)
			if err != nil {
				if attempt == 2 {
					http.Error(w, "Failed to publish tweet", http.StatusInternalServerError)
					return
				}
				time.Sleep(2 * time.Second)
				continue
			}

			// If we get a 500 error, retry once
			if resp.StatusCode == 500 && attempt == 1 {
				fmt.Printf("DEBUG: Twitter API 500 error, retrying...\n")
				resp.Body.Close()
				time.Sleep(2 * time.Second)
				continue
			}

			break
		}
		defer resp.Body.Close()

		// Handle response
		if resp.StatusCode == http.StatusCreated {
			// Success - Tweet posted
			var twitterResp TwitterPostResponse
			if err := json.NewDecoder(resp.Body).Decode(&twitterResp); err != nil {
				// Even if we can't decode response, the tweet was posted successfully
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Tweet published successfully"))
				return
			}

			// Return success with tweet ID
			response := map[string]interface{}{
				"message": "Tweet published successfully",
				"tweetId": twitterResp.Data.ID,
				"text":    twitterResp.Data.Text,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Handle errors
		fmt.Printf("DEBUG: Twitter API response status: %d\n", resp.StatusCode)

		// Read the full response body for debugging
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), resp.StatusCode)
			return
		}

		fmt.Printf("DEBUG: Twitter API response body: %s\n", string(bodyBytes))

		// Handle specific error cases
		if resp.StatusCode == 500 {
			// Twitter API internal server error - this is usually temporary
			http.Error(w, "Twitter API is experiencing temporary issues. Please try again in a few minutes.", http.StatusServiceUnavailable)
			return
		}

		if resp.StatusCode == 429 {
			// Rate limiting
			http.Error(w, "Twitter API rate limit exceeded. Please wait a moment before trying again.", http.StatusTooManyRequests)
			return
		}

		if resp.StatusCode == 400 {
			// Check if it's a Cloudflare response
			if strings.Contains(string(bodyBytes), "cloudflare") || strings.Contains(string(bodyBytes), "400 Bad Request") {
				http.Error(w, "Twitter API request was blocked. This might be due to rate limiting or temporary issues. Please try again later.", http.StatusBadRequest)
				return
			}
		}

		// Check for specific error patterns
		if strings.Contains(string(bodyBytes), "The string did not match the expected pattern") {
			// This error often occurs when the text format is invalid
			http.Error(w, "Invalid tweet text format. Please check for special characters or formatting issues.", http.StatusBadRequest)
			return
		}

		var errorResp TwitterErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err != nil {
			http.Error(w, fmt.Sprintf("Twitter API error (status: %d, body: %s)", resp.StatusCode, string(bodyBytes)), resp.StatusCode)
			return
		}

		// Return specific error message
		if len(errorResp.Errors) > 0 {
			errorMsg := errorResp.Errors[0].Message
			http.Error(w, fmt.Sprintf("Twitter API error: %s", errorMsg), resp.StatusCode)
			return
		}

		http.Error(w, "Unknown Twitter API error", resp.StatusCode)
	}
}

// GetTwitterPostsHandler fetches the user's recent tweets with metrics and media
// func GetTwitterPostsHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		userID, err := middleware.GetUserIDFromContext(r)
// 		if err != nil {
// 			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
// 			return
// 		}

// 		// Get Twitter access token
// 		var accessToken string
// 		var tokenExpiry *time.Time
// 		var refreshToken *string
// 		err = db.QueryRow(`
// 			SELECT access_token, access_token_expires_at, refresh_token
// 			FROM social_accounts
// 			WHERE user_id = $1 AND platform = 'twitter'
// 		`, userID).Scan(&accessToken, &tokenExpiry, &refreshToken)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				http.Error(w, "Twitter account not connected", http.StatusBadRequest)
// 				return
// 			}
// 			http.Error(w, "Failed to retrieve Twitter account", http.StatusInternalServerError)
// 			return
// 		}
// 		if tokenExpiry != nil && time.Now().After(*tokenExpiry) {
// 			http.Error(w, "Twitter access token has expired. Please reconnect your account.", http.StatusUnauthorized)
// 			return
// 		}

// 		// Step 1: Get the user's Twitter ID
// 		userURL := "https://api.twitter.com/2/users/me?user.fields=id,name,username,profile_image_url"
// 		reqUser, err := http.NewRequest("GET", userURL, nil)
// 		if err != nil {
// 			http.Error(w, "Failed to create request to Twitter API", http.StatusInternalServerError)
// 			return
// 		}
// 		reqUser.Header.Set("Authorization", "Bearer "+accessToken)
// 		client := &http.Client{Timeout: 30 * time.Second}
// 		respUser, err := client.Do(reqUser)
// 		if err != nil {
// 			http.Error(w, "Failed to contact Twitter API", http.StatusInternalServerError)
// 			return
// 		}
// 		defer respUser.Body.Close()
// 		if respUser.StatusCode != http.StatusOK {
// 			body, _ := io.ReadAll(respUser.Body)
// 			http.Error(w, "Failed to get Twitter user: "+string(body), respUser.StatusCode)
// 			return
// 		}
// 		var userResp struct {
// 			Data struct {
// 				ID              string `json:"id"`
// 				Name            string `json:"name"`
// 				Username        string `json:"username"`
// 				ProfileImageURL string `json:"profile_image_url"`
// 			} `json:"data"`
// 		}
// 		if err := json.NewDecoder(respUser.Body).Decode(&userResp); err != nil {
// 			http.Error(w, "Failed to decode Twitter user response", http.StatusInternalServerError)
// 			return
// 		}
// 		if userResp.Data.ID == "" {
// 			http.Error(w, "Could not get Twitter user ID", http.StatusInternalServerError)
// 			return
// 		}

// 		// Step 2: Fetch tweets with metrics and media
// 		tweetsURL := fmt.Sprintf("https://api.twitter.com/2/users/%s/tweets?max_results=20&tweet.fields=public_metrics,created_at,attachments&expansions=attachments.media_keys&media.fields=preview_image_url,url,type", userResp.Data.ID)
// 		reqTweets, err := http.NewRequest("GET", tweetsURL, nil)
// 		if err != nil {
// 			http.Error(w, "Failed to create request to Twitter API", http.StatusInternalServerError)
// 			return
// 		}
// 		reqTweets.Header.Set("Authorization", "Bearer "+accessToken)
// 		respTweets, err := client.Do(reqTweets)
// 		if err != nil {
// 			http.Error(w, "Failed to contact Twitter API", http.StatusInternalServerError)
// 			return
// 		}
// 		defer respTweets.Body.Close()
// 		if respTweets.StatusCode != http.StatusOK {
// 			body, _ := io.ReadAll(respTweets.Body)
// 			http.Error(w, "Failed to fetch Twitter posts: "+string(body), respTweets.StatusCode)
// 			return
// 		}
// 		var tweetsResp map[string]interface{}
// 		if err := json.NewDecoder(respTweets.Body).Decode(&tweetsResp); err != nil {
// 			http.Error(w, "Failed to decode Twitter posts", http.StatusInternalServerError)
// 			return
// 		}
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(tweetsResp)
// 	}
// }
