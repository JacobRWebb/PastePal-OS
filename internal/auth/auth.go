package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/JacobRWebb/PastePal-OS/internal/crypto"
	"github.com/JacobRWebb/PastePal-OS/internal/models"
	"golang.org/x/crypto/pbkdf2"
)

// HashPasswordForServer creates a hash that can be sent to the server for authentication
// This hash doesn't expose the actual master password
func HashPasswordForServer(password, email string) (string, error) {
	// Use a different salt derivation than the master key
	salt := fmt.Sprintf("pastepal-auth:%s", email)
	authHash := pbkdf2.Key(
		[]byte(password),
		[]byte(salt),
		100000,
		32,
		sha256.New,
	)

	return base64.StdEncoding.EncodeToString(authHash), nil
}

// RegisterUser prepares user registration data for the server
func RegisterUser(email, password string) (*models.RegistrationData, error) {
	// Generate master key from password and email
	masterKey, err := crypto.DeriveKeyFromPassword(password, email)
	if err != nil {
		return nil, err
	}

	// Generate a random symmetric key for file encryption
	symmetricKey, err := crypto.GenerateSymmetricKey()
	if err != nil {
		return nil, err
	}

	// Encrypt the symmetric key with the master key
	encryptedSymmetricKey, err := crypto.EncryptSymmetricKey(symmetricKey, masterKey)
	if err != nil {
		return nil, err
	}

	// Hash the password for server authentication
	passwordHash, err := HashPasswordForServer(password, email)
	if err != nil {
		return nil, err
	}

	return &models.RegistrationData{
		Email:                email,
		PasswordHash:         passwordHash,
		EncryptedSymmetricKey: encryptedSymmetricKey,
	}, nil
}

// LoginUser authenticates a user and returns the decrypted symmetric key
func LoginUser(email, password string, encryptedSymmetricKey string) ([]byte, error) {
	// Derive master key from password and email
	masterKey, err := crypto.DeriveKeyFromPassword(password, email)
	if err != nil {
		return nil, err
	}

	// Decrypt the symmetric key
	symmetricKey, err := crypto.DecryptSymmetricKey(encryptedSymmetricKey, masterKey)
	if err != nil {
		return nil, errors.New("invalid credentials or corrupted key")
	}

	return symmetricKey, nil
}

// PrepareLoginRequest creates the authentication data to send to the server
func PrepareLoginRequest(email, password string) (*models.LoginRequest, error) {
	// Hash the password for server authentication
	passwordHash, err := HashPasswordForServer(password, email)
	if err != nil {
		return nil, err
	}

	return &models.LoginRequest{
		Email:        email,
		PasswordHash: passwordHash,
	}, nil
}
