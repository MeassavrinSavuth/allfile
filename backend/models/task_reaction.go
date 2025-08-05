package models

import "time"

type TaskReaction struct {
	ID           string    `json:"id"`
	TaskID       string    `json:"task_id"`
	UserID       string    `json:"user_id"`
	ReactionType string    `json:"reaction_type"` // e.g., "thumbsUp", "heart", "laugh", etc.
	CreatedAt    time.Time `json:"created_at"`
}

// CREATE TABLE task_reactions (
//   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//   task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
//   user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//   reaction_type TEXT NOT NULL,
//   created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
//   UNIQUE(task_id, user_id, reaction_type)
// );
