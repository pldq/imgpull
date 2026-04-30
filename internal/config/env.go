package config

import "os"

// LoadEnv loads environment variables into config
func (c *Config) LoadEnv() {
	c.GitHubToken = os.Getenv("IMGPULL_GITHUB_TOKEN")
	c.GitHubRepo = os.Getenv("IMGPULL_GITHUB_REPO")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.GitHubToken == "" {
		return ErrGitHubTokenNotSet
	}
	if c.GitHubRepo == "" {
		return ErrGitHubRepoNotSet
	}
	return nil
}
