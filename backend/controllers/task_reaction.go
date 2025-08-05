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

// ToggleReaction toggles a reaction on a task (adds if not exists, removes if exists)
func ToggleReaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	var req struct {
		ReactionType string `json:"reaction_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.ReactionType == "" {
		http.Error(w, "Reaction type is required", http.StatusBadRequest)
		return
	}


	// Check if reaction already exists
	var existingReactionID string
	err := lib.DB.QueryRow(`
		SELECT id FROM task_reactions 
		WHERE task_id = $1 AND user_id = $2 AND reaction_type = $3
	`, taskID, userID, req.ReactionType).Scan(&existingReactionID)

	if err == nil {
		// Reaction exists, remove it
		_, err = lib.DB.Exec(`DELETE FROM task_reactions WHERE id = $1`, existingReactionID)
		if err != nil {
			log.Printf("[ERROR] Failed to remove reaction: %v", err)
			http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"action":        "removed",
			"reaction_type": req.ReactionType,
		})
	} else {
		// Reaction doesn't exist, add it
		reactionID := uuid.NewString()
		now := time.Now()


		_, err = lib.DB.Exec(`
			INSERT INTO task_reactions (id, task_id, user_id, reaction_type, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, reactionID, taskID, userID, req.ReactionType, now)
		if err != nil {
			log.Printf("[ERROR] Failed to add reaction: %v", err)
			http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
			return
		}

		reaction := models.TaskReaction{
			ID:           reactionID,
			TaskID:       taskID,
			UserID:       userID,
			ReactionType: req.ReactionType,
			CreatedAt:    now,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"action":   "added",
			"reaction": reaction,
		})
	}
}

// GetTaskReactions gets all reactions for a task with counts
func GetTaskReactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]


	// Get reaction counts grouped by type
	rows, err := lib.DB.Query(`
		SELECT reaction_type, COUNT(*) as count
		FROM task_reactions 
		WHERE task_id = $1 
		GROUP BY reaction_type
	`, taskID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch reactions: %v", err)
		http.Error(w, "Failed to fetch reactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	reactions := make(map[string]int)
	for rows.Next() {
		var reactionType string
		var count int
		err := rows.Scan(&reactionType, &count)
		if err != nil {
			log.Printf("[ERROR] Failed to scan reaction row: %v", err)
			continue
		}
		reactions[reactionType] = count
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactions)
}

// GetUserReactions gets all reactions by the current user for a task
func GetUserReactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	taskID := vars["taskId"]


	rows, err := lib.DB.Query(`
		SELECT reaction_type
		FROM task_reactions 
		WHERE task_id = $1 AND user_id = $2
	`, taskID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch user reactions: %v", err)
		http.Error(w, "Failed to fetch user reactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	userReactions := []string{}
	for rows.Next() {
		var reactionType string
		err := rows.Scan(&reactionType)
		if err != nil {
			log.Printf("[ERROR] Failed to scan user reaction row: %v", err)
			continue
		}
		userReactions = append(userReactions, reactionType)
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userReactions)
}
