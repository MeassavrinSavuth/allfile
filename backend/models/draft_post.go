package models

import "time"

// DraftPost represents a social media post draft
// CREATE TABLE draft_posts (
//
//	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
//	created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
//	content TEXT,
//	media JSONB,
//	platforms TEXT[],
//	status TEXT NOT NULL DEFAULT 'draft',
//	scheduled_time TIMESTAMP WITH TIME ZONE,
//	published_time TIMESTAMP WITH TIME ZONE,
//	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
//	updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
//
// );
type DraftPost struct {
	ID            string     `json:"id"`
	WorkspaceID   string     `json:"workspace_id"`
	CreatedBy     string     `json:"created_by"`
	Content       string     `json:"content"`
	Media         []string   `json:"media"` // or []Media if you want richer objects
	Platforms     []string   `json:"platforms"`
	Status        string     `json:"status"` // draft, scheduled, published
	ScheduledTime *time.Time `json:"scheduled_time"`
	PublishedTime *time.Time `json:"published_time"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
