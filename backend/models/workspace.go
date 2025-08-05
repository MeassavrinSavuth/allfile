package models

import "time"

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Avatar    *string   `json:"avatar"`
	AdminID   string    `json:"admin_id"`
	AdminName string    `json:"admin_name"`
	CreatedAt time.Time `json:"created_at"`
}

type WorkspaceMember struct {
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}
