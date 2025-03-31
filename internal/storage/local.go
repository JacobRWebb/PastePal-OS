package storage

import (
	"encoding/json"
	"errors"
	"github.com/JacobRWebb/PastePal-OS/internal/models"
	"os"
	"path/filepath"
	"sync"
)

// LocalStorage handles local storage of user data
type LocalStorage struct {
	basePath string
	mutex    sync.RWMutex
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Ensure the base directory exists
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, err
	}

	return &LocalStorage{
		basePath: basePath,
		mutex:    sync.RWMutex{},
	}, nil
}

// SaveUserSession saves user session data locally
func (ls *LocalStorage) SaveUserSession(email string, symmetricKey []byte) error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	// Create user directory if it doesn't exist
	userDir := filepath.Join(ls.basePath, "users")
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}

	// Save symmetric key to memory only (not to disk for security)
	// In a real application, you might want to encrypt this with a device-specific key
	// and store it securely, possibly using OS-specific secure storage APIs
	
	// For now, we'll just save the email to a session file
	sessionData := map[string]string{
		"email": email,
	}

	data, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(userDir, "session.json"), data, 0600)
}

// SaveCredentials saves user credentials for the remember me feature
func (ls *LocalStorage) SaveCredentials(email, passwordHash string, rememberMe bool) error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	// Create user directory if it doesn't exist
	userDir := filepath.Join(ls.basePath, "users")
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}

	// Only save credentials if remember me is enabled
	if !rememberMe {
		// If remember me is disabled, remove any existing credentials
		credPath := filepath.Join(userDir, "credentials.json")
		if _, err := os.Stat(credPath); err == nil {
			return os.Remove(credPath)
		}
		return nil
	}

	// Save credentials
	credentialsData := map[string]string{
		"email":        email,
		"passwordHash": passwordHash,
	}

	data, err := json.Marshal(credentialsData)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(userDir, "credentials.json"), data, 0600)
}

// GetSavedCredentials retrieves saved credentials if they exist
func (ls *LocalStorage) GetSavedCredentials() (email, passwordHash string, err error) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()

	credPath := filepath.Join(ls.basePath, "users", "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return "", "", errors.New("no saved credentials")
	}

	var credentialsData map[string]string
	if err := json.Unmarshal(data, &credentialsData); err != nil {
		return "", "", err
	}

	email, ok := credentialsData["email"]
	if !ok {
		return "", "", errors.New("invalid credentials data: missing email")
	}

	passwordHash, ok = credentialsData["passwordHash"]
	if !ok {
		return "", "", errors.New("invalid credentials data: missing password hash")
	}

	return email, passwordHash, nil
}

// GetUserSession retrieves the current user session
func (ls *LocalStorage) GetUserSession() (string, error) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()

	sessionPath := filepath.Join(ls.basePath, "users", "session.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return "", errors.New("no active session")
	}

	var sessionData map[string]string
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return "", err
	}

	email, ok := sessionData["email"]
	if !ok {
		return "", errors.New("invalid session data")
	}

	return email, nil
}

// ClearUserSession removes the current user session
func (ls *LocalStorage) ClearUserSession() error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	sessionPath := filepath.Join(ls.basePath, "users", "session.json")
	// Remove the session file if it exists
	if _, err := os.Stat(sessionPath); err == nil {
		return os.Remove(sessionPath)
	}

	return nil
}

// SavePasteLocally saves a paste to local storage
func (ls *LocalStorage) SavePasteLocally(paste *models.Paste) error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	// Create pastes directory if it doesn't exist
	pastesDir := filepath.Join(ls.basePath, "pastes")
	if err := os.MkdirAll(pastesDir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(paste)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(pastesDir, paste.ID+".json"), data, 0600)
}

// GetLocalPastes retrieves all locally saved pastes
func (ls *LocalStorage) GetLocalPastes() ([]*models.Paste, error) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()

	pastesDir := filepath.Join(ls.basePath, "pastes")
	if _, err := os.Stat(pastesDir); os.IsNotExist(err) {
		return []*models.Paste{}, nil
	}

	files, err := os.ReadDir(pastesDir)
	if err != nil {
		return nil, err
	}

	pastes := make([]*models.Paste, 0, len(files))
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(pastesDir, file.Name()))
		if err != nil {
			continue
		}

		var paste models.Paste
		if err := json.Unmarshal(data, &paste); err != nil {
			continue
		}

		pastes = append(pastes, &paste)
	}

	return pastes, nil
}
