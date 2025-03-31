package models

import "time"

// User represents a user in the system
type User struct {
	ID                   string    `json:"id"`
	Email                string    `json:"email"`
	EncryptedSymmetricKey string    `json:"encrypted_symmetric_key"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
}

// RegistrationData contains the data needed to register a user
type RegistrationData struct {
	Email                string `json:"email"`
	PasswordHash         string `json:"password_hash"`
	EncryptedSymmetricKey string `json:"encrypted_symmetric_key"`
}

// LoginRequest contains the data needed to authenticate a user
type LoginRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

// UserResponse represents the user object in server responses
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginResponse contains the server's response to a login request
type LoginResponse struct {
	User                 UserResponse `json:"user"`
	AuthToken            string       `json:"auth_token"`
	EncryptedSymmetricKey string       `json:"encrypted_symmetric_key"`
	// Keep these for backward compatibility
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}
