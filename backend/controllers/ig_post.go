package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"social-sync-backend/middleware" // Assuming this path is correct for your project
)

// Define the Instagram Graph API version to use
const instagramAPIVersion = "v20.0" // <--- IMPORTANT: Update to the latest stable version

type InstagramPostRequest struct {
	Caption   string   `json:"caption"`
	MediaUrls []string `json:"mediaUrls"`
}

// waitForMediaReady polls Instagram media container status until ready or timeout
// This function checks the status of individual media containers (images/videos)
// and also the carousel container itself.
func waitForMediaReady(mediaID, accessToken string) error {
	// Construct the URL to check the media container's status
	statusURL := fmt.Sprintf("https://graph.facebook.com/%s/%s?fields=status_code&access_token=%s", instagramAPIVersion, mediaID, accessToken)

	const maxRetries = 30                // Increased retries for more robust waiting (from 10)
	const delay = 5 * time.Second        // Increased delay (from 3s), total wait time now up to 150 seconds
	const initialDelay = 3 * time.Second // Initial delay before the first retry

	// Small initial delay before starting the loop to allow Instagram some initial processing time
	time.Sleep(initialDelay)

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(statusURL)
		if err != nil {
			return fmt.Errorf("failed to get media status: %w", err) // Return immediately on network error
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read media status response: %w", err) // Return immediately on read error
		}

		if resp.StatusCode != http.StatusOK {
			// Check if it's a specific Instagram error that means it will never be ready
			var errRes struct {
				Error struct {
					Message string `json:"message"`
					Code    int    `json:"code"`
					Type    string `json:"type"`
				} `json:"error"`
			}
			if json.Unmarshal(body, &errRes) == nil {
				if errRes.Error.Code == 100 && strings.Contains(strings.ToLower(errRes.Error.Message), "invalid parameter") {
					// This often indicates a permanent issue with the media itself (e.g., corrupted, unsupported format)
					return fmt.Errorf("media processing failed due to invalid media content: %s", errRes.Error.Message)
				}
			}
			return fmt.Errorf("media status check failed with HTTP status %d: %s", resp.StatusCode, body)
		}

		var res struct {
			StatusCode string `json:"status_code"`
		}
		if err := json.Unmarshal(body, &res); err != nil {
			return fmt.Errorf("failed to parse media status response: %w", err)
		}

		if res.StatusCode == "FINISHED" {
			return nil // Media is ready to publish
		} else if res.StatusCode == "ERROR" {
			return fmt.Errorf("media upload failed with status 'ERROR'")
		}

		// Media is not yet finished, wait and retry
		time.Sleep(delay)
	}

	// If loop finishes, it means media wasn't ready within the max retries
	return fmt.Errorf("media not ready for ID %s after %d retries (%s total wait)", mediaID, maxRetries, time.Duration(maxRetries)*delay)
}

func PostToInstagramHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		var req InstagramPostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(req.Caption) == "" {
			http.Error(w, "Caption cannot be empty", http.StatusBadRequest)
			return
		}

		mediaCount := len(req.MediaUrls)
		if mediaCount == 0 {
			http.Error(w, "Instagram requires at least one media URL", http.StatusBadRequest)
			return
		}
		if mediaCount > 10 {
			http.Error(w, "Instagram carousel posts can have at most 10 media items", http.StatusBadRequest)
			return
		}

		var accessToken, instagramUserID string
		err = db.QueryRow(`
			SELECT access_token, social_id
			FROM social_accounts
			WHERE user_id = $1 AND platform = 'instagram'`, userID).Scan(&accessToken, &instagramUserID)
		if err != nil {
			http.Error(w, "Instagram account not connected", http.StatusBadRequest)
			return
		}

		mediaContainerIDs := make([]string, 0, mediaCount)

		// Declare 'body' here once for the entire function scope
		var body []byte

		for _, mediaURL := range req.MediaUrls {
			form := url.Values{}
			form.Set("is_carousel_item", "true") // All media items are considered carousel items for this flow

			lower := strings.ToLower(mediaURL)
			if strings.HasSuffix(lower, ".mp4") || strings.HasSuffix(lower, ".mov") {
				form.Set("media_type", "VIDEO")
				form.Set("video_url", mediaURL)
			} else {
				form.Set("media_type", "IMAGE")
				form.Set("image_url", mediaURL)
			}

			form.Set("access_token", accessToken)

			// Step 1: Create individual media container
			createMediaURL := fmt.Sprintf("https://graph.facebook.com/%s/%s/media", instagramAPIVersion, instagramUserID)
			resp, err := http.Post(
				createMediaURL,
				"application/x-www-form-urlencoded",
				strings.NewReader(form.Encode()),
			)
			if err != nil {
				http.Error(w, "Failed to create media container", http.StatusInternalServerError)
				return
			}
			// Use '=' for reassignment, not ':='
			body, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				http.Error(w, "Failed to read media container creation response", http.StatusInternalServerError)
				return
			}

			if resp.StatusCode != http.StatusOK {
				http.Error(w, fmt.Sprintf("Media container creation failed: %s", body), http.StatusInternalServerError)
				return
			}

			var result struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(body, &result); err != nil || result.ID == "" {
				http.Error(w, "Invalid response from media container creation", http.StatusInternalServerError)
				return
			}

			// Step 2: Wait for individual media container to be ready
			if err := waitForMediaReady(result.ID, accessToken); err != nil {
				http.Error(w, fmt.Sprintf("Media item failed to process: %v", err), http.StatusInternalServerError)
				return
			}

			mediaContainerIDs = append(mediaContainerIDs, result.ID)
		}

		if mediaCount == 1 {
			// Single media post publish (This path handles both images and videos)
			publishForm := url.Values{}
			publishForm.Set("creation_id", mediaContainerIDs[0])
			publishForm.Set("caption", req.Caption) // Add caption for single media posts
			publishForm.Set("access_token", accessToken)

			publishURL := fmt.Sprintf("https://graph.facebook.com/%s/%s/media_publish", instagramAPIVersion, instagramUserID)
			publishResp, err := http.Post(publishURL, "application/x-www-form-urlencoded", strings.NewReader(publishForm.Encode()))
			if err != nil {
				http.Error(w, "Failed to publish post", http.StatusInternalServerError)
				return
			}
			// Use '=' for reassignment, not ':='
			body, err = io.ReadAll(publishResp.Body)
			publishResp.Body.Close()
			if err != nil {
				http.Error(w, "Failed to read publish response", http.StatusInternalServerError)
				return
			}

			if publishResp.StatusCode != http.StatusOK {
				http.Error(w, fmt.Sprintf("Publish failed: %s", body), http.StatusInternalServerError)
				return
			}

		} else {
			// Carousel post creation and publish
			carouselForm := url.Values{}
			carouselForm.Set("media_type", "CAROUSEL")
			carouselForm.Set("children", strings.Join(mediaContainerIDs, ","))
			carouselForm.Set("caption", req.Caption)
			carouselForm.Set("access_token", accessToken)

			// Step 3: Create carousel container
			createCarouselURL := fmt.Sprintf("https://graph.facebook.com/%s/%s/media", instagramAPIVersion, instagramUserID)
			carouselResp, err := http.Post(createCarouselURL, "application/x-www-form-urlencoded", strings.NewReader(carouselForm.Encode()))
			if err != nil {
				http.Error(w, "Failed to create carousel container", http.StatusInternalServerError)
				return
			}
			// Use '=' for reassignment, not ':='
			body, err = io.ReadAll(carouselResp.Body)
			carouselResp.Body.Close()
			if err != nil {
				http.Error(w, "Failed to read carousel container creation response", http.StatusInternalServerError)
				return
			}

			if carouselResp.StatusCode != http.StatusOK {
				http.Error(w, fmt.Sprintf("Carousel container creation failed: %s", body), http.StatusInternalServerError)
				return
			}

			var carouselResult struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(body, &carouselResult); err != nil || carouselResult.ID == "" {
				http.Error(w, "Invalid carousel container creation response", http.StatusInternalServerError)
				return
			}

			// Step 4: Wait for carousel container to be ready
			if err := waitForMediaReady(carouselResult.ID, accessToken); err != nil {
				http.Error(w, fmt.Sprintf("Carousel post failed to process: %v", err), http.StatusInternalServerError)
				return
			}

			// Step 5: Publish the carousel
			publishForm := url.Values{}
			publishForm.Set("creation_id", carouselResult.ID)
			publishForm.Set("access_token", accessToken)

			publishURL := fmt.Sprintf("https://graph.facebook.com/%s/%s/media_publish", instagramAPIVersion, instagramUserID)
			publishResp, err := http.Post(publishURL, "application/x-www-form-urlencoded", strings.NewReader(publishForm.Encode()))
			if err != nil {
				http.Error(w, "Failed to publish carousel post", http.StatusInternalServerError)
				return
			}
			// Use '=' for reassignment, not ':='
			body, err = io.ReadAll(publishResp.Body)
			publishResp.Body.Close()
			if err != nil {
				http.Error(w, "Failed to read publish response", http.StatusInternalServerError)
				return
			}

			if publishResp.StatusCode != http.StatusOK {
				http.Error(w, fmt.Sprintf("Carousel publish failed: %s", body), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Instagram post published successfully"))
	}
}

// GetInstagramPostsHandler fetches the user's Instagram posts
func GetInstagramPostsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// Get Instagram access token
		var accessToken string
		err = db.QueryRow(`
			SELECT access_token
			FROM social_accounts
			WHERE user_id = $1 AND platform = 'instagram'
		`, userID).Scan(&accessToken)
		if err == sql.ErrNoRows {
			http.Error(w, "Instagram account not connected", http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, "Failed to get Instagram account", http.StatusInternalServerError)
			return
		}

		// Fetch posts from Instagram Graph API
		graphURL := "https://graph.instagram.com/me/media?fields=id,caption,media_type,media_url,permalink,thumbnail_url,timestamp,like_count,comments_count&access_token=" + accessToken
		resp, err := http.Get(graphURL)
		if err != nil {
			http.Error(w, "Failed to contact Instagram API", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			http.Error(w, "Failed to fetch Instagram posts: "+string(body), resp.StatusCode)
			return
		}
		var igResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&igResp); err != nil {
			http.Error(w, "Failed to decode Instagram posts", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(igResp)
	}
}
