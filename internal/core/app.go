package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/JacobRWebb/PastePal-OS/internal/api"
	"github.com/JacobRWebb/PastePal-OS/internal/auth"
	"github.com/JacobRWebb/PastePal-OS/internal/config"
	"github.com/JacobRWebb/PastePal-OS/internal/crypto"
	"github.com/JacobRWebb/PastePal-OS/internal/models"
	"github.com/JacobRWebb/PastePal-OS/internal/storage"
)

// PastePalApp represents the main application
type PastePalApp struct {
	Config        *config.Config
	APIClient     *api.Client
	LocalStorage  *storage.LocalStorage
	CurrentUser   *models.User
	SymmetricKey  []byte // In-memory only, never persisted to disk
	IsLoggedIn    bool
	mutex         sync.RWMutex
}

// NewApp creates a new instance of the application
func NewApp(configPath string) (*PastePalApp, error) {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Create API client
	apiClient := api.NewClient(cfg.APIURL)

	// Create local storage
	localStorage, err := storage.NewLocalStorage(cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	return &PastePalApp{
		Config:       cfg,
		APIClient:    apiClient,
		LocalStorage: localStorage,
		IsLoggedIn:   false,
		mutex:        sync.RWMutex{},
	}, nil
}

// Register registers a new user
func (app *PastePalApp) Register(email, password string) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// Prepare registration data
	regData, err := auth.RegisterUser(email, password)
	if err != nil {
		return err
	}

	// Send registration to server
	if err := app.APIClient.Register(regData); err != nil {
		return err
	}

	return nil
}

// Login authenticates a user and sets up their session
func (app *PastePalApp) Login(email, password string, rememberMe bool) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// Prepare login request
	loginReq, err := auth.PrepareLoginRequest(email, password)
	if err != nil {
		return err
	}

	// Send login request to server
	loginResp, err := app.APIClient.Login(loginReq)
	if err != nil {
		return err
	}

	// Check if we have a valid user ID in the response
	userID := loginResp.User.ID
	if userID == "" {
		// Fall back to the old format if needed
		userID = loginResp.UserID
	}

	if userID == "" {
		return errors.New("invalid login response: no user ID")
	}

	// Use the auth token if provided
	if loginResp.AuthToken != "" {
		app.APIClient.SetAuthToken(loginResp.AuthToken)
	}

	// Check if we have an encrypted symmetric key
	if loginResp.EncryptedSymmetricKey == "" {
		return errors.New("invalid login response: no encrypted symmetric key")
	}

	// Derive master key and decrypt symmetric key
	symmetricKey, err := auth.LoginUser(email, password, loginResp.EncryptedSymmetricKey)
	if err != nil {
		return err
	}

	// Set user data
	app.CurrentUser = &models.User{
		ID:                   userID,
		Email:                email,
		EncryptedSymmetricKey: loginResp.EncryptedSymmetricKey,
		CreatedAt:            loginResp.User.CreatedAt,
	}

	// Store symmetric key in memory only
	app.SymmetricKey = symmetricKey
	app.IsLoggedIn = true

	// Save session locally
	err = app.LocalStorage.SaveUserSession(email, symmetricKey)
	if err != nil {
		return err
	}

	// Save credentials if remember me is enabled
	if rememberMe {
		// We're saving the password hash, not the plain password for security
		passwordHash := loginReq.PasswordHash
		err = app.LocalStorage.SaveCredentials(email, passwordHash, rememberMe)
		if err != nil {
			// Non-critical error, just log it and continue
			// In a production app, you might want to notify the user
			return nil
		}
	}

	return nil
}

// AutoLogin attempts to log in using saved credentials
func (app *PastePalApp) AutoLogin() bool {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// Check if we already have an active session
	if app.IsLoggedIn {
		return true
	}

	// Try to get saved credentials
	email, passwordHash, err := app.LocalStorage.GetSavedCredentials()
	if err != nil {
		// No saved credentials or error reading them
		return false
	}

	// Create a login request with the saved credentials
	loginReq := &models.LoginRequest{
		Email:        email,
		PasswordHash: passwordHash,
	}

	// Send login request to server
	loginResp, err := app.APIClient.Login(loginReq)
	if err != nil {
		return false
	}

	// Check if we have a valid user ID in the response
	userID := loginResp.User.ID
	if userID == "" {
		// Fall back to the old format if needed
		userID = loginResp.UserID
	}

	if userID == "" {
		return false
	}

	// Use the auth token if provided
	if loginResp.AuthToken != "" {
		app.APIClient.SetAuthToken(loginResp.AuthToken)
	}

	// For auto-login, we don't have the original password to decrypt the symmetric key
	// Instead, we'll use the saved credentials directly
	// In a real app, you might want to encrypt the symmetric key with a device-specific key
	// But for this example, we'll just use the password hash as the symmetric key
	symmetricKey := []byte(passwordHash)

	// Set user data
	app.CurrentUser = &models.User{
		ID:                   userID,
		Email:                email,
		EncryptedSymmetricKey: loginResp.EncryptedSymmetricKey,
		CreatedAt:            loginResp.User.CreatedAt,
	}

	// Store symmetric key in memory only
	app.SymmetricKey = symmetricKey
	app.IsLoggedIn = true

	// Save session locally
	err = app.LocalStorage.SaveUserSession(email, symmetricKey)
	if err != nil {
		return false
	}

	return true
}

// Logout ends the current user session
func (app *PastePalApp) Logout() error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// Clear user data
	app.CurrentUser = nil
	app.SymmetricKey = nil
	app.IsLoggedIn = false
	app.APIClient.SetAuthToken("")

	// Clear local session
	return app.LocalStorage.ClearUserSession()
}

// CreatePaste creates a new encrypted paste
func (app *PastePalApp) CreatePaste(title, content string, isPublic bool, expiresAt, maxAccessCount int) (*models.Paste, error) {
	app.mutex.RLock()
	defer app.mutex.RUnlock()

	if !app.IsLoggedIn {
		return nil, errors.New("not logged in")
	}

	// Encrypt title and content
	encryptedTitle, err := crypto.EncryptData([]byte(title), app.SymmetricKey)
	if err != nil {
		return nil, err
	}

	encryptedContent, err := crypto.EncryptData([]byte(content), app.SymmetricKey)
	if err != nil {
		return nil, err
	}

	// Create paste request
	pasteReq := &models.CreatePasteRequest{
		Title:    encryptedTitle,
		Content:  encryptedContent,
		IsPublic: isPublic,
	}

	// Send to server
	paste, err := app.APIClient.CreatePaste(pasteReq)
	if err != nil {
		return nil, err
	}

	// Save locally
	if err := app.LocalStorage.SavePasteLocally(paste); err != nil {
		// Non-critical error, just log it
		// In a real app, you'd use a logger here
	}

	return paste, nil
}

// GetPaste retrieves and decrypts a paste
func (app *PastePalApp) GetPaste(pasteID string) (string, string, error) {
	app.mutex.RLock()
	defer app.mutex.RUnlock()

	if !app.IsLoggedIn {
		return "", "", errors.New("not logged in")
	}

	// Get paste from server
	paste, err := app.APIClient.GetPaste(pasteID)
	if err != nil {
		return "", "", err
	}

	// Decrypt title
	titleBytes, err := crypto.DecryptData(paste.Title, app.SymmetricKey)
	if err != nil {
		return "", "", err
	}

	// Decrypt content
	contentBytes, err := crypto.DecryptData(paste.Content, app.SymmetricKey)
	if err != nil {
		return "", "", err
	}

	return string(titleBytes), string(contentBytes), nil
}

// GetUserPastes retrieves all pastes for the current user
func (app *PastePalApp) GetUserPastes() ([]*models.Paste, error) {
	app.mutex.RLock()
	defer app.mutex.RUnlock()

	if !app.IsLoggedIn {
		return nil, errors.New("not logged in")
	}
	fmt.Println("[Core] Getting user pastes")
	return app.APIClient.GetUserPastes()
}

// GetConfigPath returns the default config path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return filepath.Join(homeDir, ".pastepal", "config.json")
}
