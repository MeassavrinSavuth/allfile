package models

import (
	"time"

	"github.com/lib/pq"
)

// Media represents uploaded media files in workspaces
// CREATE TABLE media (
//
//	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
//	uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
//	filename TEXT NOT NULL,
//	original_name TEXT NOT NULL,
//	file_url TEXT NOT NULL,
//	file_type TEXT NOT NULL, -- 'image' or 'video'
//	mime_type TEXT NOT NULL,
//	file_size INTEGER NOT NULL,
//	width INTEGER, -- for images/videos
//	height INTEGER, -- for images/videos
//	duration FLOAT, -- for videos (in seconds)
//	tags TEXT[],
//	cloudinary_public_id TEXT,
//	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
//	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
//
// );
type Media struct {
	ID                 string         `json:"id"`
	WorkspaceID        string         `json:"workspace_id"`
	UploadedBy         string         `json:"uploaded_by"`
	Filename           string         `json:"filename"`
	OriginalName       string         `json:"original_name"`
	FileURL            string         `json:"file_url"`
	FileType           string         `json:"file_type"` // 'image' or 'video'
	MimeType           string         `json:"mime_type"`
	FileSize           int64          `json:"file_size"`
	Width              *int           `json:"width,omitempty"`
	Height             *int           `json:"height,omitempty"`
	Duration           *float64       `json:"duration,omitempty"` // for videos
	Tags               pq.StringArray `json:"tags" gorm:"type:text[]"`
	CloudinaryPublicID string         `json:"cloudinary_public_id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`

	// Joined fields
	UploaderName   string `json:"uploader_name,omitempty"`
	UploaderAvatar string `json:"uploader_avatar,omitempty"`
}
