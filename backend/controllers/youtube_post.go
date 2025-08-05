package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"social-sync-backend/lib"
	"strings"
	"time"

	"social-sync-backend/middleware"

	"golang.org/x/oauth2"
)

// YouTubeUploadRequest represents the video upload request
// (renamed to YouTubeVideoMetadata)
type YouTubeVideoMetadata struct {
	Snippet struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags,omitempty"`
		CategoryID  string   `json:"categoryId"`
	} `json:"snippet"`
	Status struct {
		PrivacyStatus string `json:"privacyStatus"`
	} `json:"status"`
}

// YouTubeUploadResponse represents YouTube API response
type YouTubeUploadResponse struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		PublishedAt  string                 `json:"publishedAt"`
		ChannelID    string                 `json:"channelId"`
		Title        string                 `json:"title"`
		Description  string                 `json:"description"`
		Thumbnails   map[string]interface{} `json:"thumbnails"`
		ChannelTitle string                 `json:"channelTitle"`
		Tags         []string               `json:"tags"`
		CategoryID   string                 `json:"categoryId"`
	} `json:"snippet"`
	Status struct {
		UploadStatus        string `json:"uploadStatus"`
		PrivacyStatus       string `json:"privacyStatus"`
		License             string `json:"license"`
		Embeddable          bool   `json:"embeddable"`
		PublicStatsViewable bool   `json:"publicStatsViewable"`
	} `json:"status"`
}

// PostToYouTubeHandler handles video upload to YouTube
func PostToYouTubeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "user not authenticated", http.StatusUnauthorized)
			return
		}

		err = r.ParseMultipartForm(100 << 20)
		if err != nil {
			http.Error(w, "failed to parse form data", http.StatusBadRequest)
			return
		}

		file, fileHeader, err := r.FormFile("video")
		if err != nil {
			http.Error(w, "video file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if !isValidVideoFile(fileHeader.Filename) {
			http.Error(w, "invalid video file format. supported: mp4, mov, avi, wmv, flv, webm, mkv", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		description := r.FormValue("description")
		tags := r.FormValue("tags")
		privacy := r.FormValue("privacy")
		categoryID := r.FormValue("category_id")

		if title == "" {
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}

		if privacy == "" {
			privacy = "private"
		}
		if categoryID == "" {
			categoryID = "22"
		}

		var accessToken, refreshToken string
		err = db.QueryRow(`
			SELECT access_token, refresh_token 
			FROM social_accounts 
			WHERE user_id = $1 AND platform = 'youtube'
		`, userID).Scan(&accessToken, &refreshToken)

		if err == sql.ErrNoRows {
			http.Error(w, "YouTube account not connected", http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, "failed to get YouTube account", http.StatusInternalServerError)
			return
		}

		cloudinaryURL, err := lib.UploadToCloudinary(file, "videos", fileHeader.Filename)
		if err != nil {
			http.Error(w, "failed to upload video to storage", http.StatusInternalServerError)
			return
		}

		file.Seek(0, 0)

		videoID, err := uploadVideoToYouTube(file, title, description, tags, privacy, categoryID, accessToken)
		if err != nil {
			if strings.Contains(err.Error(), "401") && refreshToken != "" {
				newAccessToken, err := refreshYouTubeToken(refreshToken)
				if err != nil {
					http.Error(w, "failed to refresh YouTube token", http.StatusUnauthorized)
					return
				}

				_, err = db.Exec(`
					UPDATE social_accounts 
					SET access_token = $1, last_synced_at = $2
					WHERE user_id = $3 AND platform = 'youtube'
				`, newAccessToken, time.Now(), userID)
				if err != nil {
					http.Error(w, "failed to update access token", http.StatusInternalServerError)
					return
				}

				file.Seek(0, 0)
				videoID, err = uploadVideoToYouTube(file, title, description, tags, privacy, categoryID, newAccessToken)
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to upload to YouTube: %v", err), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, fmt.Sprintf("failed to upload to YouTube: %v", err), http.StatusInternalServerError)
				return
			}
		}

		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

		response := map[string]interface{}{
			"message":    "video uploaded successfully to YouTube",
			"video_id":   videoID,
			"video_url":  videoURL,
			"backup_url": cloudinaryURL,
			"title":      title,
			"privacy":    privacy,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func uploadVideoToYouTube(file multipart.File, title, description, tags, privacy, categoryID, accessToken string) (string, error) {
	metadata := YouTubeVideoMetadata{}
	metadata.Snippet.Title = title
	metadata.Snippet.Description = description
	metadata.Snippet.CategoryID = categoryID
	metadata.Status.PrivacyStatus = privacy

	if tags != "" {
		tagList := strings.Split(tags, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
		metadata.Snippet.Tags = tagList
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: Sending metadata to YouTube: %s\n", string(metadataJSON))

	initURL := "https://www.googleapis.com/upload/youtube/v3/videos?uploadType=resumable&part=snippet,status"
	initReq, err := http.NewRequest("POST", initURL, bytes.NewReader(metadataJSON))
	if err != nil {
		return "", err
	}

	initReq.Header.Set("Authorization", "Bearer "+accessToken)
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("X-Upload-Content-Type", "video/*")
	initReq.Header.Set("X-Upload-Content-Length", "0")

	client := &http.Client{Timeout: 30 * time.Second}
	initResp, err := client.Do(initReq)
	if err != nil {
		return "", err
	}
	defer initResp.Body.Close()

	fmt.Printf("DEBUG: Init response status: %d\n", initResp.StatusCode)

	if initResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(initResp.Body)
		fmt.Printf("DEBUG: Init error response: %s\n", string(bodyBytes))
		return "", fmt.Errorf("failed to initialize upload: %d - %s", initResp.StatusCode, string(bodyBytes))
	}

	uploadURL := initResp.Header.Get("Location")
	if uploadURL == "" {
		return "", fmt.Errorf("no upload URL received from YouTube")
	}

	fmt.Printf("DEBUG: Upload URL: %s\n", uploadURL)

	file.Seek(0, 0)

	uploadReq, err := http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		return "", err
	}

	uploadReq.Header.Set("Content-Type", "video/*")
	uploadReq.Header.Set("Content-Length", "0")

	uploadClient := &http.Client{Timeout: 300 * time.Second}
	uploadResp, err := uploadClient.Do(uploadReq)
	if err != nil {
		return "", err
	}
	defer uploadResp.Body.Close()

	fmt.Printf("DEBUG: Upload response status: %d\n", uploadResp.StatusCode)

	if uploadResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(uploadResp.Body)
		fmt.Printf("DEBUG: Upload error response: %s\n", string(bodyBytes))
		return "", fmt.Errorf("failed to upload video: %d - %s", uploadResp.StatusCode, string(bodyBytes))
	}

	var uploadRespData YouTubeUploadResponse
	if err := json.NewDecoder(uploadResp.Body).Decode(&uploadRespData); err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: YouTube response - Title: %s, Privacy: %s\n", uploadRespData.Snippet.Title, uploadRespData.Status.PrivacyStatus)
	fmt.Printf("DEBUG: Full YouTube response: %+v\n", uploadRespData)

	return uploadRespData.ID, nil
}

func refreshYouTubeToken(refreshToken string) (string, error) {
	config := getYouTubeOAuthConfig()
	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := config.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", err
	}
	return newToken.AccessToken, nil
}

func isValidVideoFile(filename string) bool {
	validExtensions := []string{".mp4", ".mov", ".avi", ".wmv", ".flv", ".webm", ".mkv"}
	filename = strings.ToLower(filename)
	for _, ext := range validExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// GetYouTubePostsHandler fetches the user's YouTube videos with stats
func GetYouTubePostsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.GetUserIDFromContext(r)
		if err != nil {
			http.Error(w, "user not authenticated", http.StatusUnauthorized)
			return
		}

		// Get YouTube access token
		var accessToken string
		var refreshToken string
		err = db.QueryRow(`
			SELECT access_token, refresh_token 
			FROM social_accounts 
			WHERE user_id = $1 AND platform = 'youtube'
		`, userID).Scan(&accessToken, &refreshToken)
		if err == sql.ErrNoRows {
			http.Error(w, "YouTube account not connected", http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, "failed to get YouTube account", http.StatusInternalServerError)
			return
		}

		// Step 1: Get the user's channel ID
		channelsURL := "https://www.googleapis.com/youtube/v3/channels?part=id,snippet&mine=true"
		req, err := http.NewRequest("GET", channelsURL, nil)
		if err != nil {
			http.Error(w, "failed to create request to YouTube API", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "failed to contact YouTube API", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			http.Error(w, "failed to get YouTube channel: "+string(body), resp.StatusCode)
			return
		}
		var channelsResp struct {
			Items []struct {
				ID      string `json:"id"`
				Snippet struct {
					Title       string                 `json:"title"`
					Description string                 `json:"description"`
					Thumbnails  map[string]interface{} `json:"thumbnails"`
				} `json:"snippet"`
			} `json:"items"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&channelsResp); err != nil {
			http.Error(w, "failed to decode YouTube channel response", http.StatusInternalServerError)
			return
		}
		if len(channelsResp.Items) == 0 {
			http.Error(w, "No YouTube channel found", http.StatusBadRequest)
			return
		}
		channelID := channelsResp.Items[0].ID

		// Step 2: Get the user's videos (playlistItems for uploads)
		playlistURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=contentDetails&id=%s", channelID)
		req2, err := http.NewRequest("GET", playlistURL, nil)
		if err != nil {
			http.Error(w, "failed to create request to YouTube API", http.StatusInternalServerError)
			return
		}
		req2.Header.Set("Authorization", "Bearer "+accessToken)
		resp2, err := client.Do(req2)
		if err != nil {
			http.Error(w, "failed to contact YouTube API", http.StatusInternalServerError)
			return
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp2.Body)
			http.Error(w, "failed to get YouTube uploads playlist: "+string(body), resp2.StatusCode)
			return
		}
		var playlistResp struct {
			Items []struct {
				ContentDetails struct {
					RelatedPlaylists struct {
						Uploads string `json:"uploads"`
					} `json:"relatedPlaylists"`
				} `json:"contentDetails"`
			} `json:"items"`
		}
		if err := json.NewDecoder(resp2.Body).Decode(&playlistResp); err != nil {
			http.Error(w, "failed to decode YouTube playlist response", http.StatusInternalServerError)
			return
		}
		if len(playlistResp.Items) == 0 {
			http.Error(w, "No uploads playlist found", http.StatusBadRequest)
			return
		}
		uploadsPlaylistID := playlistResp.Items[0].ContentDetails.RelatedPlaylists.Uploads

		// Step 3: Get videos from the uploads playlist
		videosURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&maxResults=20&playlistId=%s", uploadsPlaylistID)
		req3, err := http.NewRequest("GET", videosURL, nil)
		if err != nil {
			http.Error(w, "failed to create request to YouTube API", http.StatusInternalServerError)
			return
		}
		req3.Header.Set("Authorization", "Bearer "+accessToken)
		resp3, err := client.Do(req3)
		if err != nil {
			http.Error(w, "failed to contact YouTube API", http.StatusInternalServerError)
			return
		}
		defer resp3.Body.Close()
		if resp3.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp3.Body)
			http.Error(w, "failed to get YouTube videos: "+string(body), resp3.StatusCode)
			return
		}
		var videosResp struct {
			Items []struct {
				Snippet struct {
					Title       string                 `json:"title"`
					Description string                 `json:"description"`
					Thumbnails  map[string]interface{} `json:"thumbnails"`
					PublishedAt string                 `json:"publishedAt"`
					ResourceID  struct {
						VideoID string `json:"videoId"`
					} `json:"resourceId"`
				} `json:"snippet"`
				ContentDetails struct {
					VideoID string `json:"videoId"`
				} `json:"contentDetails"`
			} `json:"items"`
		}
		if err := json.NewDecoder(resp3.Body).Decode(&videosResp); err != nil {
			http.Error(w, "failed to decode YouTube videos response", http.StatusInternalServerError)
			return
		}

		// Step 4: Get video statistics for each video
		videoIDs := []string{}
		for _, item := range videosResp.Items {
			videoIDs = append(videoIDs, item.Snippet.ResourceID.VideoID)
		}
		if len(videoIDs) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		statsURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=statistics,snippet&id=%s", strings.Join(videoIDs, ","))
		req4, err := http.NewRequest("GET", statsURL, nil)
		if err != nil {
			http.Error(w, "failed to create request to YouTube API", http.StatusInternalServerError)
			return
		}
		req4.Header.Set("Authorization", "Bearer "+accessToken)
		resp4, err := client.Do(req4)
		if err != nil {
			http.Error(w, "failed to contact YouTube API", http.StatusInternalServerError)
			return
		}
		defer resp4.Body.Close()
		if resp4.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp4.Body)
			http.Error(w, "failed to get YouTube video stats: "+string(body), resp4.StatusCode)
			return
		}
		var statsResp map[string]interface{}
		if err := json.NewDecoder(resp4.Body).Decode(&statsResp); err != nil {
			http.Error(w, "failed to decode YouTube video stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statsResp)
	}
}
