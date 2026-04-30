package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDirName = "imgpull"
)

// Errors
var (
	ErrGitHubTokenNotSet = fmt.Errorf("IMGPULL_GITHUB_TOKEN environment variable is not set")
	ErrGitHubRepoNotSet  = fmt.Errorf("IMGPULL_GITHUB_REPO environment variable is not set")
)

// Load loads configuration from database and environment variables
func Load(dbPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Load from database if exists
	if dbPath != "" {
		if err := cfg.loadFromDB(dbPath); err != nil {
			// Ignore error if database doesn't exist yet
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load config from database: %w", err)
			}
		}
	}

	cfg.LoadEnv()
	return cfg, nil
}

// Save saves the configuration to database
func (c *Config) Save(dbPath string) error {
	return c.saveToDB(dbPath)
}

// GetDBPath returns the default path to the SQLite database
func GetDBPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "images.db"), nil
}

func getConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, appDirName), nil
}
