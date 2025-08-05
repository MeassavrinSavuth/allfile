package models

import "time"

type WorkspaceInvitation struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	Email         string    `json:"email"`
	InviterID     string    `json:"inviter_id"`
	InviterName   string    `json:"inviter_name"`
	WorkspaceName string    `json:"workspace_name"`
	Status        string    `json:"status"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}
