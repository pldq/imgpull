package config

import (
	"image-poller/internal/db"
)

// DockerMode represents the docker connection mode
type DockerMode string

const (
	DockerModeCLI DockerMode = "cli"
	DockerModeAPI DockerMode = "api"
)

// Config holds the application configuration
type Config struct {
	DockerMode  DockerMode `json:"docker_mode"`
	DockerHost  string     `json:"docker_host"`
	GitHubToken string     `json:"-"` // Loaded from env, not persisted
	GitHubRepo  string     `json:"-"` // Loaded from env, not persisted
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DockerMode: DockerModeCLI,
		DockerHost: "",
	}
}

// loadFromDB loads config from database
func (c *Config) loadFromDB(dbPath string) error {
	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer database.Close()

	// Load docker_mode
	mode, err := database.GetConfig("docker_mode")
	if err != nil {
		return err
	}
	if mode != "" {
		c.DockerMode = DockerMode(mode)
	}

	// Load docker_host
	host, err := database.GetConfig("docker_host")
	if err != nil {
		return err
	}
	c.DockerHost = host

	return nil
}

// saveToDB saves config to database
func (c *Config) saveToDB(dbPath string) error {
	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := database.SetConfig("docker_mode", string(c.DockerMode)); err != nil {
		return err
	}

	if err := database.SetConfig("docker_host", c.DockerHost); err != nil {
		return err
	}

	return nil
}
