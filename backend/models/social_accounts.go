// models/social_account.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// SocialAccount struct matches your PostgreSQL table schema
type SocialAccount struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"userId"` // Corresponds to your app's user_id
	Platform             string     `json:"platform"`
	SocialID             string     `json:"socialId"` // The platform's user ID (e.g., Facebook ID)
	AccessToken          string     `json:"-"`        // Don't expose token to frontend
	AccessTokenExpiresAt *time.Time `json:"accessTokenExpiresAt"`
	RefreshToken         *string    `json:"-"`        // Don't expose token to frontend
	ProfilePictureURL    *string    `json:"profilePictureUrl"` // Pointers for nullable fields
	ProfileName          *string    `json:"profileName"`       // Pointers for nullable fields
	ConnectedAt          time.Time  `json:"connectedAt"`
	LastSyncedAt         *time.Time `json:"lastSyncedAt"`
}