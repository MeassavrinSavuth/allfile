package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"social-sync-backend/lib"
	"social-sync-backend/middleware"
	"social-sync-backend/models"
	"time"

	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Mock DB for demonstration (replace with real DB logic)
var workspaces = []models.Workspace{}

// --- WebSocket Real-Time Hub ---
type wsClient struct {
	conn *websocket.Conn
}

type wsHub struct {
	clients map[string]map[*wsClient]bool // workspaceId -> clients
	lock    sync.RWMutex
}

var hub = &wsHub{
	clients: make(map[string]map[*wsClient]bool),
}

func (h *wsHub) addClient(workspaceId string, client *wsClient) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.clients[workspaceId] == nil {
		h.clients[workspaceId] = make(map[*wsClient]bool)
	}
	h.clients[workspaceId][client] = true
}

func (h *wsHub) removeClient(workspaceId string, client *wsClient) {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.clients[workspaceId] != nil {
		delete(h.clients[workspaceId], client)
		if len(h.clients[workspaceId]) == 0 {
			delete(h.clients, workspaceId)
		}
	}
}

func (h *wsHub) broadcast(workspaceId string, messageType int, data []byte) {
	h.lock.RLock()
	defer h.lock.RUnlock()
	for client := range h.clients[workspaceId] {
		client.conn.WriteMessage(messageType, data)
	}
}

// WebSocket handler for workspace real-time updates
func WorkspaceWSHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceId := vars["workspaceId"]
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &wsClient{conn: conn}
	hub.addClient(workspaceId, client)
	defer func() {
		hub.removeClient(workspaceId, client)
		conn.Close()
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// This server is broadcast-only; ignore client messages
	}
}

func ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// TODO: Filter by user (currently returns all workspaces)
	rows, err := lib.DB.Query(`
		SELECT w.id, w.name, w.avatar, w.admin_id, u.name as admin_name, w.created_at
		FROM workspaces w
		INNER JOIN workspace_members wm ON w.id = wm.workspace_id
		INNER JOIN users u ON w.admin_id = u.id
		WHERE wm.user_id = $1
		ORDER BY w.created_at DESC
	`, userID)
	if err != nil {
		log.Printf("Error fetching workspaces: %v", err)
		http.Error(w, "Failed to fetch workspaces", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var workspaces []models.Workspace
	for rows.Next() {
		var ws models.Workspace
		var adminName *string
		err := rows.Scan(&ws.ID, &ws.Name, &ws.Avatar, &ws.AdminID, &adminName, &ws.CreatedAt)
		if err != nil {
			log.Printf("Error scanning workspace: %v", err)
			http.Error(w, "Failed to process workspace data", http.StatusInternalServerError)
			return
		}
		if adminName != nil {
			ws.AdminName = *adminName
		} else {
			ws.AdminName = "Unknown User"
		}
		workspaces = append(workspaces, ws)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating workspaces: %v", err)
		http.Error(w, "Failed to process workspaces", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaces)
}

func CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req struct {
		Name   string  `json:"name"`
		Avatar *string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Workspace name is required", http.StatusBadRequest)
		return
	}

	workspaceID := uuid.NewString()
	now := time.Now()

	// Insert workspace into database
	_, err := lib.DB.Exec(`
		INSERT INTO workspaces (id, name, avatar, admin_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, workspaceID, req.Name, req.Avatar, userID, now)
	if err != nil {
		log.Printf("Error creating workspace: %v", err)
		http.Error(w, "Failed to create workspace", http.StatusInternalServerError)
		return
	}

	// Create workspace member record for admin
	_, err = lib.DB.Exec(`
		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
	`, workspaceID, userID, "Admin", now)
	if err != nil {
		log.Printf("Error creating workspace member: %v", err)
		http.Error(w, "Failed to create workspace member", http.StatusInternalServerError)
		return
	}

	// Get user name for response
	var adminName *string
	err = lib.DB.QueryRow("SELECT name FROM users WHERE id = $1", userID).Scan(&adminName)
	if err != nil {
		log.Printf("Error fetching user name: %v", err)
		// Continue without name if there's an error
	}

	ws := models.Workspace{
		ID:      workspaceID,
		Name:    req.Name,
		Avatar:  req.Avatar,
		AdminID: userID,
		AdminName: func() string {
			if adminName != nil {
				return *adminName
			}
			return "Unknown User"
		}(),
		CreatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ws)
}

func ListWorkspaceMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	rows, err := lib.DB.Query(`
		SELECT u.id, u.name, u.email, u.profile_picture, wm.role
		FROM workspace_members wm
		INNER JOIN users u ON wm.user_id = u.id
		WHERE wm.workspace_id = $1
		ORDER BY wm.role DESC, u.name ASC
	`, workspaceID)
	if err != nil {
		http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Member struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		Email  string  `json:"email"`
		Avatar *string `json:"avatar"`
		Role   string  `json:"role"`
	}
	var members []Member
	for rows.Next() {
		var m Member
		err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Avatar, &m.Role)
		if err != nil {
			continue
		}
		members = append(members, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// LeaveWorkspace allows a user to leave a workspace
func LeaveWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// Get workspace ID from URL
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	// Check if user is the admin of the workspace
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

	if isAdmin {
		http.Error(w, "Admins cannot leave their own workspace. Transfer ownership first.", http.StatusForbidden)
		return
	}

	// Remove user from workspace
	result, err := lib.DB.Exec(`
		DELETE FROM workspace_members 
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID)

	if err != nil {
		log.Printf("Error removing user from workspace: %v", err)
		http.Error(w, "Failed to leave workspace", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User is not a member of this workspace", http.StatusNotFound)
		return
	}

	// Get user's email
	var userEmail string
	err = lib.DB.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&userEmail)
	if err == nil && userEmail != "" {
		msg, _ := json.Marshal(map[string]interface{}{
			"type":         "removed_from_workspace",
			"workspace_id": workspaceID,
		})
		userHub.broadcast(userEmail, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully left workspace"})
}

// RemoveWorkspaceMember allows an admin to remove a member from the workspace
func RemoveWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT token
	userID := r.Context().Value(middleware.UserIDKey).(string)

	// Get workspace ID and member ID from URL
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]
	memberID := vars["memberId"]

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to verify permissions"})
		return
	}

	if !isAdmin {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only workspace admins can remove members"})
		return
	}

	// Check if trying to remove self (admin)
	if userID == memberID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Admins cannot remove themselves. Use leave workspace instead."})
		return
	}

	// Check if member exists and is not admin
	var memberRole string
	err = lib.DB.QueryRow(`
		SELECT role FROM workspace_members 
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, memberID).Scan(&memberRole)

	if err != nil {
		log.Printf("Error checking member: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Member not found in workspace"})
		return
	}

	if memberRole == "Admin" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cannot remove another admin"})
		return
	}

	// Remove member from workspace
	result, err := lib.DB.Exec(`
		DELETE FROM workspace_members 
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, memberID)

	if err != nil {
		log.Printf("Error removing member from workspace: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to remove member"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Member not found in workspace"})
		return
	}

	// Get member's email
	var memberEmail string
	err = lib.DB.QueryRow("SELECT email FROM users WHERE id = $1", memberID).Scan(&memberEmail)
	if err == nil && memberEmail != "" {
		msg, _ := json.Marshal(map[string]interface{}{
			"type":         "removed_from_workspace",
			"workspace_id": workspaceID,
		})
		userHub.broadcast(memberEmail, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Member removed successfully"})

	// --- WebSocket broadcast for real-time member removal ---
	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "member_removed",
		"user_id": memberID,
	})
	hub.broadcast(workspaceID, websocket.TextMessage, msg)
}

// DeleteWorkspace allows the admin to delete a workspace and all related data
func DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	// Check if user is admin of the workspace
	var isAdmin bool
	err := lib.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM workspaces WHERE id = $1 AND admin_id = $2
		)
	`, workspaceID, userID).Scan(&isAdmin)
	if err != nil {
		log.Printf("Error checking admin status: %v", err)
		http.Error(w, "Failed to verify permissions", http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Only the workspace admin can delete this workspace", http.StatusForbidden)
		return
	}

	// Start transaction
	tx, err := lib.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Failed to delete workspace", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete workspace members
	_, err = tx.Exec(`DELETE FROM workspace_members WHERE workspace_id = $1`, workspaceID)
	if err != nil {
		log.Printf("Error deleting workspace members: %v", err)
		http.Error(w, "Failed to delete workspace members", http.StatusInternalServerError)
		return
	}

	// Delete workspace invitations
	_, err = tx.Exec(`DELETE FROM workspace_invitations WHERE workspace_id = $1`, workspaceID)
	if err != nil {
		log.Printf("Error deleting workspace invitations: %v", err)
		http.Error(w, "Failed to delete workspace invitations", http.StatusInternalServerError)
		return
	}

	// Delete the workspace itself
	_, err = tx.Exec(`DELETE FROM workspaces WHERE id = $1`, workspaceID)
	if err != nil {
		log.Printf("Error deleting workspace: %v", err)
		http.Error(w, "Failed to delete workspace", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to delete workspace", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Workspace deleted successfully"})
}

// ChangeMemberRole allows the admin to change a member's role in the workspace
func ChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]
	memberID := vars["memberId"]

	// Only admin can change roles
	var isAdmin bool
	err := lib.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2 AND role = 'Admin'
		)
	`, workspaceID, userID).Scan(&isAdmin)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to verify permissions"})
		return
	}
	if !isAdmin {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only workspace admin can change member roles"})
		return
	}

	// Prevent admin from changing their own role
	if userID == memberID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Admin cannot change their own role"})
		return
	}

	// Parse new role from request body
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}
	if req.Role != "Admin" && req.Role != "Editor" && req.Role != "Viewer" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid role"})
		return
	}

	// Update the member's role
	result, err := lib.DB.Exec(`
		UPDATE workspace_members SET role = $1 WHERE workspace_id = $2 AND user_id = $3
	`, req.Role, workspaceID, memberID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update member role"})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Member not found in workspace", http.StatusNotFound)
		return
	}

	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "member_role_changed",
		"user_id": memberID,
		"role":    req.Role,
	})
	hub.broadcast(workspaceID, websocket.TextMessage, msg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Member role updated successfully"})
}
