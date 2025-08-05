package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"social-sync-backend/lib"
	"social-sync-backend/middleware"
	"social-sync-backend/models"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lib/pq"
)

// CreateDraftPost creates a new draft post
func CreateDraftPost(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset body for decoder

	var req struct {
		Content       string     `json:"content"`
		Media         []string   `json:"media"`
		Platforms     []string   `json:"platforms"`
		ScheduledTime *time.Time `json:"scheduled_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	draftID := uuid.NewString()
	now := time.Now()
	status := "draft"

	_, err := lib.DB.Exec(`
		INSERT INTO draft_posts (id, workspace_id, created_by, content, media, platforms, status, scheduled_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, draftID, workspaceID, userID, req.Content, pqStringArrayToJSONB(req.Media), pqStringArray(req.Platforms), status, req.ScheduledTime, now, now)
	if err != nil {
		http.Error(w, "Failed to create draft", http.StatusInternalServerError)
		return
	}

	// Fetch author info
	var authorName, authorEmail, authorAvatar string
	err = lib.DB.QueryRow(`SELECT name, email, profile_picture FROM users WHERE id = $1`, userID).Scan(&authorName, &authorEmail, &authorAvatar)
	if err != nil {
		authorName = "Unknown"
		authorEmail = ""
		authorAvatar = "/default-avatar.png"
	}

	response := map[string]interface{}{
		"id":             draftID,
		"workspace_id":   workspaceID,
		"created_by":     userID,
		"content":        req.Content,
		"media":          req.Media,
		"platforms":      req.Platforms,
		"status":         status,
		"scheduled_time": req.ScheduledTime,
		"created_at":     now,
		"updated_at":     now,
		"author": map[string]interface{}{
			"id":     userID,
			"name":   authorName,
			"email":  authorEmail,
			"avatar": authorAvatar,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	msg, _ := json.Marshal(map[string]interface{}{
		"type":  "draft_created",
		"draft": response,
	})
	hub.broadcast(workspaceID, websocket.TextMessage, msg)
}

// ListDraftPosts lists all draft posts for a workspace
func ListDraftPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspaceID := vars["workspaceId"]

	rows, err := lib.DB.Query(`
		SELECT d.id, d.workspace_id, d.created_by, d.content, d.media, d.platforms, d.status, d.scheduled_time, d.published_time, d.created_at, d.updated_at,
		       u.id, u.name, u.email, u.profile_picture
		FROM draft_posts d
		LEFT JOIN users u ON d.created_by = u.id
		WHERE d.workspace_id = $1 ORDER BY d.created_at DESC
	`, workspaceID)
	if err != nil {
		http.Error(w, "Failed to fetch drafts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Author struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar"`
	}

	var drafts []map[string]interface{}
	for rows.Next() {
		var d models.DraftPost
		var mediaJSON []byte
		var platforms pqStringArray
		var authorID, authorName, authorEmail, authorAvatar *string
		if err := rows.Scan(&d.ID, &d.WorkspaceID, &d.CreatedBy, &d.Content, &mediaJSON, &platforms, &d.Status, &d.ScheduledTime, &d.PublishedTime, &d.CreatedAt, &d.UpdatedAt, &authorID, &authorName, &authorEmail, &authorAvatar); err != nil {
			continue
		}
		d.Media = jsonBytesToStringSlice(mediaJSON)
		d.Platforms = []string(platforms)
		m := map[string]interface{}{
			"id":             d.ID,
			"workspace_id":   d.WorkspaceID,
			"created_by":     d.CreatedBy,
			"content":        d.Content,
			"media":          d.Media,
			"platforms":      d.Platforms,
			"status":         d.Status,
			"scheduled_time": d.ScheduledTime,
			"published_time": d.PublishedTime,
			"created_at":     d.CreatedAt,
			"updated_at":     d.UpdatedAt,
		}
		if authorID != nil {
			m["author"] = Author{
				ID:     *authorID,
				Name:   authorNameOrEmail(authorName, authorEmail),
				Email:  authorEmailOrEmpty(authorEmail),
				Avatar: authorAvatarOrDefault(authorAvatar),
			}
		}
		drafts = append(drafts, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drafts)
}

func authorNameOrEmail(name, email *string) string {
	if name != nil && *name != "" {
		return *name
	}
	if email != nil {
		return *email
	}
	return "Unknown"
}

func authorEmailOrEmpty(email *string) string {
	if email != nil {
		return *email
	}
	return ""
}

func authorAvatarOrDefault(avatar *string) string {
	if avatar != nil && *avatar != "" {
		return *avatar
	}
	return "/default-avatar.png"
}

// UpdateDraftPost updates a draft post
func UpdateDraftPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	draftID := vars["draftId"]

	var req struct {
		Content       *string    `json:"content"`
		Media         *[]string  `json:"media"`
		Platforms     *[]string  `json:"platforms"`
		ScheduledTime *time.Time `json:"scheduled_time"`
		Status        *string    `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1
	if req.Content != nil {
		setClauses = append(setClauses, "content = $"+itoa(argIdx))
		args = append(args, *req.Content)
		argIdx++
	}
	if req.Media != nil {
		setClauses = append(setClauses, "media = $"+itoa(argIdx))
		args = append(args, pqStringArrayToJSONB(*req.Media))
		argIdx++
	}
	if req.Platforms != nil {
		setClauses = append(setClauses, "platforms = $"+itoa(argIdx))
		args = append(args, pqStringArray(*req.Platforms))
		argIdx++
	}
	if req.ScheduledTime != nil {
		setClauses = append(setClauses, "scheduled_time = $"+itoa(argIdx))
		args = append(args, *req.ScheduledTime)
		argIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, "status = $"+itoa(argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	setClauses = append(setClauses, "updated_at = $"+itoa(argIdx))
	args = append(args, time.Now())
	argIdx++
	if len(setClauses) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}
	args = append(args, draftID)
	query := "UPDATE draft_posts SET " + joinClauses(setClauses, ", ") + " WHERE id = $" + itoa(argIdx)
	_, err := lib.DB.Exec(query, args...)
	if err != nil {
		http.Error(w, "Failed to update draft", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Draft updated successfully"})

	msg, _ := json.Marshal(map[string]interface{}{
		"type":  "draft_updated",
		"draft": req,
	})
	hub.broadcast(vars["workspaceId"], websocket.TextMessage, msg)
}

// DeleteDraftPost deletes a draft post
func DeleteDraftPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	draftID := vars["draftId"]

	_, err := lib.DB.Exec(`DELETE FROM draft_posts WHERE id = $1`, draftID)
	if err != nil {
		http.Error(w, "Failed to delete draft", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Draft deleted successfully"})

	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "draft_deleted",
		"draftId": draftID,
	})
	hub.broadcast(vars["workspaceId"], websocket.TextMessage, msg)
}

// PublishDraftPost publishes a draft post (only for admin/editor)
func PublishDraftPost(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	draftID := vars["draftId"]

	// Check if user is admin/editor in the workspace
	var workspaceID string
	err := lib.DB.QueryRow(`SELECT workspace_id FROM draft_posts WHERE id = $1`, draftID).Scan(&workspaceID)
	if err != nil {
		http.Error(w, "Draft not found", http.StatusNotFound)
		return
	}
	if !IsUserAdminOrEditor(userID, workspaceID) {
		http.Error(w, "Not authorized to publish", http.StatusForbidden)
		return
	}

	now := time.Now()
	_, err = lib.DB.Exec(`
		UPDATE draft_posts SET status = 'published', published_time = $1, updated_at = $1 WHERE id = $2
	`, now, draftID)
	if err != nil {
		http.Error(w, "Failed to publish draft", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Draft published successfully"})

	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "draft_published",
		"draftId": draftID,
	})
	hub.broadcast(workspaceID, websocket.TextMessage, msg)
}

// --- Helpers ---

// pqStringArrayToJSONB converts a []string to JSONB for Postgres
func pqStringArrayToJSONB(arr []string) []byte {
	b, _ := json.Marshal(arr)
	return b
}

// pqStringArray converts a []string to Postgres TEXT[]
type pqStringArray []string

func (a pqStringArray) Value() (interface{}, error) {
	return pq.Array([]string(a)).Value()
}

func (a *pqStringArray) Scan(src interface{}) error {
	return pq.Array((*[]string)(a)).Scan(src)
}

// jsonBytesToStringSlice converts JSONB []byte to []string
func jsonBytesToStringSlice(b []byte) []string {
	var arr []string
	_ = json.Unmarshal(b, &arr)
	return arr
}

// IsUserAdminOrEditor checks if a user is admin or editor in a workspace
func IsUserAdminOrEditor(userID, workspaceID string) bool {
	var role string
	err := lib.DB.QueryRow(`SELECT role FROM workspace_members WHERE user_id = $1 AND workspace_id = $2`, userID, workspaceID).Scan(&role)
	if err != nil {
		return false
	}
	return role == "Admin" || role == "Editor"
}
