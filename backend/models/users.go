// package models

// import "time"

// type User struct {
// 	ID         string    `json:"id"`
// 	Name       string    `json:"name"` // Add this line
// 	Email      string    `json:"email"`
// 	Password   string    `json:"password"`
// 	CreatedAt  time.Time `json:"created_at"`
// 	UpdatedAt  time.Time `json:"updated_at"`
// 	IsVerified bool      `json:"is_verified"`
// 	IsActive   bool      `json:"is_active"`
// }

package models

import (
    "time"
    "github.com/google/uuid" // <--- 1. IMPORT THIS PACKAGE
)

type User struct {
    ID              uuid.UUID  `json:"id"`
    Name            *string    `json:"name"`           // Changed to *string (nullable)
    Email           string     `json:"email"`          // NOT NULL in DB
    Password        string     `json:"-"`              // Hidden from JSON
    Provider        *string    `json:"provider"`       // Nullable
    ProviderID      *string    `json:"-"`              // Nullable, hidden from JSON
    ProfilePicture  *string    `json:"profileImage"`   // Nullable
    CreatedAt       *time.Time `json:"created_at"`     // Changed to *time.Time (nullable)
    UpdatedAt       *time.Time `json:"updated_at"`     // Changed to *time.Time (nullable)
    IsVerified      *bool      `json:"is_verified"`    // Changed to *bool (nullable)
    IsActive        *bool      `json:"is_active"`      // Changed to *bool (nullable)
}

// CREATE EXTENSION IF NOT EXISTS "pgcrypto";

// CREATE TABLE users (
//   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
//   email TEXT UNIQUE NOT NULL CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
//   password TEXT NOT NULL,
//   created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
//   updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
//   is_verified BOOLEAN DEFAULT FALSE,
//   is_active BOOLEAN DEFAULT TRUE
// );


