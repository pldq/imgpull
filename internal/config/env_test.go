package config

import (
	"errors"
	"os"
	"testing"
)

func TestConfig_LoadEnv(t *testing.T) {
	os.Setenv("IMGPULL_GITHUB_TOKEN", "test-token")
	os.Setenv("IMGPULL_GITHUB_REPO", "owner/repo")
	defer os.Unsetenv("IMGPULL_GITHUB_TOKEN")
	defer os.Unsetenv("IMGPULL_GITHUB_REPO")

	cfg := &Config{}
	cfg.LoadEnv()

	if cfg.GitHubToken != "test-token" {
		t.Errorf("GitHubToken: got %q, want %q", cfg.GitHubToken, "test-token")
	}
	if cfg.GitHubRepo != "owner/repo" {
		t.Errorf("GitHubRepo: got %q, want %q", cfg.GitHubRepo, "owner/repo")
	}
}

func TestConfig_LoadEnv_Empty(t *testing.T) {
	os.Unsetenv("IMGPULL_GITHUB_TOKEN")
	os.Unsetenv("IMGPULL_GITHUB_REPO")

	cfg := &Config{}
	cfg.LoadEnv()

	if cfg.GitHubToken != "" {
		t.Errorf("GitHubToken should be empty, got %q", cfg.GitHubToken)
	}
	if cfg.GitHubRepo != "" {
		t.Errorf("GitHubRepo should be empty, got %q", cfg.GitHubRepo)
	}
}

func TestConfig_Validate_NoToken(t *testing.T) {
	cfg := &Config{
		GitHubToken: "",
		GitHubRepo:  "owner/repo",
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrGitHubTokenNotSet) {
		t.Errorf("Validate() error: got %v, want %v", err, ErrGitHubTokenNotSet)
	}
}

func TestConfig_Validate_NoRepo(t *testing.T) {
	cfg := &Config{
		GitHubToken: "test-token",
		GitHubRepo:  "",
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrGitHubRepoNotSet) {
		t.Errorf("Validate() error: got %v, want %v", err, ErrGitHubRepoNotSet)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DockerMode != DockerModeCLI {
		t.Errorf("Default DockerMode: got %q, want %q", cfg.DockerMode, DockerModeCLI)
	}
	if cfg.DockerHost != "" {
		t.Errorf("Default DockerHost should be empty, got %q", cfg.DockerHost)
	}
}
