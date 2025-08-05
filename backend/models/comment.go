package models

import "time"

type Comment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CREATE TABLE comments (
//   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//   task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
//   user_id UUID NOT NULL REFERENCES users(id),
//   content TEXT NOT NULL,
//   created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
// );
