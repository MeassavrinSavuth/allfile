// services/social_account_syncer.go (New file)
package utils

import (
	// "context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	// "golang.org/x/oauth2"
	"social-sync-backend/models"
	 // Assuming your models package is correctly imported
)

// FetchAndSyncFacebookProfile fetches the latest profile picture and name for a Facebook page
// and updates the database.
func FetchAndSyncFacebookProfile(db *sql.DB, account *models.SocialAccount) error {
	// Reconstruct the client using the stored access token
	// NOTE: For Facebook Page Access Tokens, they are generally long-lived,
	// but user tokens can expire. You might need to handle refresh token logic
	// if you're dealing with user access tokens directly.
	// For Page Access Tokens, they usually only expire if the user removes your app.
	// token := &oauth2.Token{
	// 	AccessToken: account.AccessToken,
	// 	Expiry:      time.Now().Add(24 * time.Hour), // Set a dummy future expiry for client if not used
	// }
	
	// Create a client with the existing token
	// This assumes you have the oauth2.Config available, which might be tricky if not in controllers.
	// For simplicity, we'll use a basic http.Client for Graph API calls since we have the token.
	// A more robust solution might involve re-initializing a stripped-down oauth2.Config
	// or making a custom transport.
    // For now, we'll simulate it, but ideally you'd use the oauth2 client for refresh capabilities.
    
    // A more direct HTTP client for making calls with existing access token:
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest("GET", fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=name,picture.type(large)", account.SocialID), nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Add("Authorization", "Bearer "+account.AccessToken)

    resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch Facebook page data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		log.Printf("Facebook API error for %s (%s): %v", account.Platform, account.SocialID, errorBody)
		return fmt.Errorf("facebook API returned status %d for %s: %v", resp.StatusCode, account.SocialID, errorBody)
	}

	var pageInfo struct {
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pageInfo); err != nil {
		return fmt.Errorf("failed to decode Facebook page info: %w", err)
	}

	newProfilePictureURL := pageInfo.Picture.Data.URL
	newProfileName := pageInfo.Name

	// Update the database if anything has changed
	stmt := `
        UPDATE social_accounts
        SET
            profile_picture_url = $1,
            profile_name = $2,
            last_synced_at = NOW()
        WHERE id = $3 AND user_id = $4
        AND (profile_picture_url IS DISTINCT FROM $1 OR profile_name IS DISTINCT FROM $2)
    `
    // Using IS DISTINCT FROM ensures an update only happens if the value actually changed,
    // which is good practice to avoid unnecessary DB writes and trigger updates.
    // However, if the old profile_picture_url was NULL, and new one is a URL, it will update.
    // If new one is NULL and old was a URL, it will update.

	_, err = db.Exec(stmt, newProfilePictureURL, newProfileName, account.ID, account.UserID)
	if err != nil {
		return fmt.Errorf("failed to update social account in DB (%s - %s): %w", account.Platform, account.SocialID, err)
	}

	log.Printf("Successfully synced Facebook profile for user %s, account %s (%s)", account.UserID, account.SocialID, account.Platform)
	return nil
}

// SyncAllSocialAccountsTask fetches and updates profile data for all connected accounts.
func SyncAllSocialAccountsTask(db *sql.DB) {
	log.Println("Starting scheduled social account sync...")

	rows, err := db.Query("SELECT id, user_id, platform, social_id, access_token FROM social_accounts")
	if err != nil {
		log.Printf("Error querying social accounts for sync: %v", err)
		return
	}
	defer rows.Close()

	var accountsToSync []models.SocialAccount
	for rows.Next() {
		var acc models.SocialAccount
		// Make sure to select all fields needed for the sync operation
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Platform, &acc.SocialID, &acc.AccessToken); err != nil {
			log.Printf("Error scanning social account row: %v", err)
			continue
		}
		accountsToSync = append(accountsToSync, acc)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating social account rows: %v", err)
		return
	}

	if len(accountsToSync) == 0 {
		log.Println("No social accounts to sync.")
		return
	}

	// Use a goroutine for each account to process in parallel, but with a limit
    // A simple semaphore can limit concurrency if you have many accounts
    maxConcurrency := 5 // Adjust based on your needs and API rate limits
    sem := make(chan struct{}, maxConcurrency)

	for _, account := range accountsToSync {
        // Handle specific platforms
        if account.Platform == "facebook" { // Ensure consistency with stored platform names
            acc := account // Capture loop variable for goroutine
            sem <- struct{}{} // Acquire a slot
            go func() {
                defer func() { <-sem }() // Release the slot
                if err := FetchAndSyncFacebookProfile(db, &acc); err != nil {
                    log.Printf("Failed to sync Facebook profile for user %s, account %s: %v", acc.UserID, acc.SocialID, err)
                }
            }()
        } else {
            // Add logic for other platforms (Instagram, YouTube etc.) here
            log.Printf("Skipping sync for unsupported platform: %s (User: %s)", account.Platform, account.UserID)
        }
	}
    // Wait for all goroutines to finish (optional, but good for clean shutdown or if you need to know when all are done)
    for i := 0; i < maxConcurrency; i++ {
        sem <- struct{}{} // Fill up the semaphore to ensure all goroutines have released their slots
    }

	log.Println("Finished scheduled social account sync.")
}