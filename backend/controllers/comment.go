package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"social-sync-backend/lib"
	"social-sync-backend/middleware"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// AddComment adds a comment to a task
func AddComment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	log.Printf("[DEBUG] Adding comment - userID: %s, taskID: %s", userID, taskID)

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Failed to decode request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		log.Printf("[ERROR] Content is empty")
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Comment content: %s", req.Content)

	commentID := uuid.NewString()
	now := time.Now()

	log.Printf("[DEBUG] Inserting comment with ID: %s", commentID)

	_, err := lib.DB.Exec(`
		INSERT INTO comments (id, task_id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, commentID, taskID, userID, req.Content, now)
	if err != nil {
		log.Printf("[ERROR] Failed to add comment to database: %v", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Fetch user name and email for response
	var userName, userEmail string
	err = lib.DB.QueryRow("SELECT name, email FROM users WHERE id = $1", userID).Scan(&userName, &userEmail)
	if err != nil {
		userName = ""
		userEmail = ""
	}
	displayName := userName
	if displayName == "" {
		displayName = userEmail
	}
	if displayName == "" {
		displayName = userID
	}

	comment := map[string]interface{}{
		"id":         commentID,
		"task_id":    taskID,
		"user_id":    userID,
		"content":    req.Content,
		"created_at": now,
		"user_name":  displayName,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// ListComments lists all comments for a task
func ListComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	log.Printf("[DEBUG] Listing comments for taskID: %s", taskID)

	// Get user name and email
	rows, err := lib.DB.Query(`
		SELECT c.id, c.task_id, c.user_id, c.content, c.created_at, u.name, u.email
		FROM comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.task_id = $1 
		ORDER BY c.created_at ASC
	`, taskID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch comments: %v", err)
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comments := []map[string]interface{}{}
	for rows.Next() {
		var id, taskID, userID, content, userName, userEmail string
		var createdAt time.Time
		err := rows.Scan(&id, &taskID, &userID, &content, &createdAt, &userName, &userEmail)
		if err != nil {
			log.Printf("[ERROR] Failed to scan comment row: %v", err)
			continue
		}

		// Use name if available, otherwise use email, fallback to user ID
		displayName := userName
		if displayName == "" {
			displayName = userEmail
		}
		if displayName == "" {
			displayName = userID
		}

		comment := map[string]interface{}{
			"id":         id,
			"task_id":    taskID,
			"user_id":    userID,
			"content":    content,
			"created_at": createdAt,
			"user_name":  displayName,
		}
		comments = append(comments, comment)
	}

	log.Printf("[DEBUG] Found %d comments for task %s", len(comments), taskID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// DeleteComment deletes a comment by ID
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	_, err := lib.DB.Exec(`DELETE FROM comments WHERE id = $1`, commentID)
	if err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Comment deleted successfully"})
}
