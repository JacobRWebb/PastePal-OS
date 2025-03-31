package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JacobRWebb/PastePal-OS/internal/models"
)

// Client handles API communication with the server
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthToken  string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetAuthToken sets the authentication token for API requests
func (c *Client) SetAuthToken(token string) {
	c.AuthToken = token
}

// Register registers a new user with the server
func (c *Client) Register(data *models.RegistrationData) error {
	reqBody, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/auth/register", c.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(body))
	}

	return nil
}

// Login authenticates a user and returns their encrypted symmetric key
func (c *Client) Login(loginReq *models.LoginRequest) (*models.LoginResponse, error) {
	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/auth/login", c.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("authentication failed")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	loginResp := &models.LoginResponse{}

	err = json.Unmarshal(respBody, loginResp)
	if err != nil {
		var rawResp map[string]interface{}
		if jsonErr := json.Unmarshal(respBody, &rawResp); jsonErr == nil {
			if userMap, ok := rawResp["user"].(map[string]interface{}); ok {
				if id, ok := userMap["id"].(string); ok {
					loginResp.User.ID = id
				}
				if email, ok := userMap["email"].(string); ok {
					loginResp.User.Email = email
				}
			}
			if token, ok := rawResp["auth_token"].(string); ok {
				loginResp.AuthToken = token
			}
			if key, ok := rawResp["encrypted_symmetric_key"].(string); ok {
				loginResp.EncryptedSymmetricKey = key
			}
		} else {
			return nil, fmt.Errorf("failed to parse server response: %v", err)
		}
	}

	if loginResp.User.ID == "" || loginResp.EncryptedSymmetricKey == "" {
		return nil, errors.New("invalid server response: missing required data")
	}

	token := resp.Header.Get("Authorization")
	if token != "" {
		c.SetAuthToken(token)
	} else if loginResp.AuthToken != "" {
		c.SetAuthToken(loginResp.AuthToken)
	}

	return loginResp, nil
}

// CreatePaste creates a new encrypted paste on the server
func (c *Client) CreatePaste(pasteReq *models.CreatePasteRequest) (*models.Paste, error) {
	if c.AuthToken == "" {
		return nil, errors.New("not authenticated")
	}

	reqBody, err := json.Marshal(pasteReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/pastes", c.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.AuthToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to create paste")
	}

	var paste models.Paste
	if err := json.NewDecoder(resp.Body).Decode(&paste); err != nil {
		return nil, err
	}

	return &paste, nil
}

// GetPaste retrieves an encrypted paste from the server
func (c *Client) GetPaste(pasteID string) (*models.Paste, error) {
	fmt.Println("[API Client] Getting paste with ID:", pasteID)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/pastes/%s", c.BaseURL, pasteID), nil)
	if err != nil {
		return nil, err
	}

	if c.AuthToken != "" {
		req.Header.Set("Authorization", c.AuthToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("paste not found or access denied")
	}

	var paste models.Paste
	if err := json.NewDecoder(resp.Body).Decode(&paste); err != nil {
		return nil, err
	}

	return &paste, nil
}

// GetUserPastes retrieves all pastes for the authenticated user
func (c *Client) GetUserPastes() ([]*models.Paste, error) {
	if c.AuthToken == "" {
		return nil, errors.New("not authenticated")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/pastes", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.AuthToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve pastes")
	}

	var pastes []*models.Paste
	if err := json.NewDecoder(resp.Body).Decode(&pastes); err != nil {
		return nil, err
	}

	return pastes, nil
}
