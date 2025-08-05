package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"social-sync-backend/lib"
	"social-sync-backend/middleware"
	"social-sync-backend/models"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// SendInvitation sends an invitation to join a workspace
func SendInvitation(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// Get workspace ID from URL
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	// Check if user is admin of the workspace
	var isAdmin bool
	err := lib.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM workspace_members 
			WHERE workspace_id = $1 AND user_id = $2 AND role = 'Admin'
		)
	`, workspaceID, userID).Scan(&isAdmin)

	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		http.Error(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}

	if !isAdmin {
		http.Error(w, "Only workspace admins can send invitations", http.StatusForbidden)
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate role
	if req.Role != "Admin" && req.Role != "Editor" && req.Role != "Viewer" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Check if invitation already exists
	var existingInvitation string
	err = lib.DB.QueryRow(`
		SELECT id FROM workspace_invitations 
		WHERE workspace_id = $1 AND email = $2 AND status = 'pending'
	`, workspaceID, req.Email).Scan(&existingInvitation)

	if err == nil {
		http.Error(w, "Invitation already sent to this email", http.StatusConflict)
		return
	}

	// Check if user is already a member
	var existingMember string
	err = lib.DB.QueryRow(`
		SELECT wm.user_id FROM workspace_members wm
		INNER JOIN users u ON wm.user_id = u.id
		WHERE wm.workspace_id = $1 AND u.email = $2
	`, workspaceID, req.Email).Scan(&existingMember)

	if err == nil {
		http.Error(w, "User is already a member of this workspace", http.StatusConflict)
		return
	}

	invitationID := uuid.NewString()
	now := time.Now()
	expiresAt := now.AddDate(0, 0, 7) // 7 days from now

	// Create invitation
	_, err = lib.DB.Exec(`
		INSERT INTO workspace_invitations (id, workspace_id, email, inviter_id, status, role, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, invitationID, workspaceID, req.Email, userID, "pending", req.Role, now, expiresAt)

	if err != nil {
		log.Printf("Error creating invitation: %v", err)
		http.Error(w, "Failed to create invitation", http.StatusInternalServerError)
		return
	}

	// Get workspace name and inviter name for response
	var workspaceName, inviterName string
	err = lib.DB.QueryRow(`
		SELECT w.name, u.name 
		FROM workspaces w, users u 
		WHERE w.id = $1 AND u.id = $2
	`, workspaceID, userID).Scan(&workspaceName, &inviterName)

	if err != nil {
		log.Printf("Error fetching workspace/inviter details: %v", err)
		// Continue without names if there's an error
	}

	invitation := models.WorkspaceInvitation{
		ID:            invitationID,
		WorkspaceID:   workspaceID,
		Email:         req.Email,
		InviterID:     userID,
		InviterName:   inviterName,
		WorkspaceName: workspaceName,
		Status:        "pending",
		Role:          req.Role,
		CreatedAt:     now,
		ExpiresAt:     expiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invitation)

	// --- WebSocket broadcast to invited user (real-time invite) ---
	msg, _ := json.Marshal(map[string]interface{}{
		"type":       "invitation_created",
		"invitation": invitation,
	})
	if userHub != nil {
		userHub.broadcast(req.Email, msg)
	}
}

// GetInvitations gets all pending invitations for the current user
func GetInvitations(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// Get user's email
	var userEmail string
	err := lib.DB.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&userEmail)
	if err != nil {
		log.Printf("Error fetching user email: %v", err)
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}

	// Get pending invitations for this email
	rows, err := lib.DB.Query(`
		SELECT wi.id, wi.workspace_id, wi.email, wi.inviter_id, wi.status, wi.role, wi.created_at, wi.expires_at,
		       w.name as workspace_name, u.name as inviter_name
		FROM workspace_invitations wi
		INNER JOIN workspaces w ON wi.workspace_id = w.id
		INNER JOIN users u ON wi.inviter_id = u.id
		WHERE wi.email = $1 AND wi.status = 'pending' AND wi.expires_at > now()
		ORDER BY wi.created_at DESC
	`, userEmail)

	if err != nil {
		log.Printf("Error fetching invitations: %v", err)
		http.Error(w, "Failed to fetch invitations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var invitations []models.WorkspaceInvitation
	for rows.Next() {
		var inv models.WorkspaceInvitation
		var inviterName *string
		err := rows.Scan(
			&inv.ID, &inv.WorkspaceID, &inv.Email, &inv.InviterID, &inv.Status, &inv.Role, &inv.CreatedAt, &inv.ExpiresAt,
			&inv.WorkspaceName, &inviterName,
		)
		if err != nil {
			log.Printf("Error scanning invitation: %v", err)
			continue
		}
		if inviterName != nil {
			inv.InviterName = *inviterName
		} else {
			inv.InviterName = "Unknown User"
		}
		invitations = append(invitations, inv)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitations)
}

// AcceptInvitation accepts a workspace invitation
func AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// Get invitation ID from URL
	vars := mux.Vars(r)
	invitationID := vars["invitationId"]

	// Get user's email
	var userEmail string
	err := lib.DB.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&userEmail)
	if err != nil {
		log.Printf("Error fetching user email: %v", err)
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}

	// Get invitation details
	var invitation models.WorkspaceInvitation
	err = lib.DB.QueryRow(`
		SELECT wi.id, wi.workspace_id, wi.email, wi.inviter_id, wi.status, wi.role, wi.created_at, wi.expires_at,
		       w.name as workspace_name, u.name as inviter_name
		FROM workspace_invitations wi
		INNER JOIN workspaces w ON wi.workspace_id = w.id
		INNER JOIN users u ON wi.inviter_id = u.id
		WHERE wi.id = $1 AND wi.email = $2 AND wi.status = 'pending' AND wi.expires_at > now()
	`, invitationID, userEmail).Scan(
		&invitation.ID, &invitation.WorkspaceID, &invitation.Email, &invitation.InviterID, &invitation.Status, &invitation.Role, &invitation.CreatedAt, &invitation.ExpiresAt,
		&invitation.WorkspaceName, &invitation.InviterName,
	)

	if err != nil {
		log.Printf("Error fetching invitation: %v", err)
		http.Error(w, "Invitation not found or expired", http.StatusNotFound)
		return
	}

	// Start transaction
	tx, err := lib.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Failed to process invitation", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update invitation status
	_, err = tx.Exec("UPDATE workspace_invitations SET status = 'accepted' WHERE id = $1", invitationID)
	if err != nil {
		log.Printf("Error updating invitation: %v", err)
		http.Error(w, "Failed to accept invitation", http.StatusInternalServerError)
		return
	}

	// Add user to workspace members
	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
	`, invitation.WorkspaceID, userID, invitation.Role, now)

	if err != nil {
		log.Printf("Error adding workspace member: %v", err)
		http.Error(w, "Failed to join workspace", http.StatusInternalServerError)
		return
	}

	// --- WebSocket broadcast for real-time member add ---
	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "member_added",
		"user_id": userID,
	})
	hub.broadcast(invitation.WorkspaceID, 1, msg)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to process invitation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation accepted successfully"})
}

// DeclineInvitation declines a workspace invitation
func DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// Get invitation ID from URL
	vars := mux.Vars(r)
	invitationID := vars["invitationId"]

	// Get user's email
	var userEmail string
	err := lib.DB.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&userEmail)
	if err != nil {
		log.Printf("Error fetching user email: %v", err)
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}

	// Update invitation status
	result, err := lib.DB.Exec(`
		UPDATE workspace_invitations 
		SET status = 'declined' 
		WHERE id = $1 AND email = $2 AND status = 'pending'
	`, invitationID, userEmail)

	if err != nil {
		log.Printf("Error declining invitation: %v", err)
		http.Error(w, "Failed to decline invitation", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Invitation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation declined successfully"})
}
