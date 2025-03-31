package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	APIURL      string `json:"api_url"`
	StoragePath string `json:"storage_path"`
	DebugMode   bool   `json:"debug_mode"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	// Get user's home directory for storage
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return &Config{
		// APIURL:      "https://api.pastepal.com",
		APIURL:      "http://localhost:8080",
		StoragePath: filepath.Join(homeDir, ".pastepal"),
		DebugMode:   false,
	}
}

// LoadConfig loads the configuration from a file
func LoadConfig(path string) (*Config, error) {
	// If the file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		config := DefaultConfig()
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
