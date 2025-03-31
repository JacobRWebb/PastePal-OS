package models

import (
	"time"
)

// Paste represents an encrypted paste stored on the server
type Paste struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Title         string    `json:"title"` // Encrypted
	Content       string    `json:"content"` // Encrypted
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	IsPublic      bool      `json:"is_public"`
	AccessCount   int       `json:"access_count,omitempty"`
	MaxAccessCount int       `json:"max_access_count,omitempty"`
}

// CreatePasteRequest represents a request to create a new paste
type CreatePasteRequest struct {
	Title         string    `json:"title"` // Already encrypted
	Content       string    `json:"content"` // Already encrypted
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	IsPublic      bool      `json:"is_public"`
	MaxAccessCount int       `json:"max_access_count,omitempty"`
}

// PasteMetadata contains non-sensitive metadata about a paste
type PasteMetadata struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	IsPublic      bool      `json:"is_public"`
	AccessCount   int       `json:"access_count,omitempty"`
	MaxAccessCount int       `json:"max_access_count,omitempty"`
}
