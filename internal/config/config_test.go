package config

import (
	"path/filepath"
	"testing"

	"image-poller/internal/db"
)

func TestConfig_loadFromDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open() error: %v", err)
	}
	database.SetConfig("docker_mode", "api")
	database.SetConfig("docker_host", "tcp://localhost:2375")
	database.Close()

	cfg := DefaultConfig()
	if err := cfg.loadFromDB(dbPath); err != nil {
		t.Errorf("loadFromDB() error: %v", err)
	}

	if cfg.DockerMode != DockerModeAPI {
		t.Errorf("DockerMode: got %q, want %q", cfg.DockerMode, DockerModeAPI)
	}
	if cfg.DockerHost != "tcp://localhost:2375" {
		t.Errorf("DockerHost: got %q, want %q", cfg.DockerHost, "tcp://localhost:2375")
	}
}

func TestConfig_loadFromDB_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := DefaultConfig()
	if err := cfg.loadFromDB(dbPath); err != nil {
		t.Errorf("loadFromDB() error: %v", err)
	}

	if cfg.DockerMode != DockerModeCLI {
		t.Errorf("DockerMode should keep default: got %q", cfg.DockerMode)
	}
	if cfg.DockerHost != "" {
		t.Errorf("DockerHost should be empty, got %q", cfg.DockerHost)
	}
}

func TestConfig_saveToDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &Config{
		DockerMode: DockerModeAPI,
		DockerHost: "tcp://localhost:2375",
	}

	if err := cfg.saveToDB(dbPath); err != nil {
		t.Errorf("saveToDB() error: %v", err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open() error: %v", err)
	}
	defer database.Close()

	mode, _ := database.GetConfig("docker_mode")
	if mode != "api" {
		t.Errorf("Saved docker_mode: got %q, want %q", mode, "api")
	}

	host, _ := database.GetConfig("docker_host")
	if host != "tcp://localhost:2375" {
		t.Errorf("Saved docker_host: got %q, want %q", host, "tcp://localhost:2375")
	}
}

func TestLoad(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Errorf("Load() error: %v", err)
	}
	if cfg.DockerMode != DockerModeCLI {
		t.Errorf("Load() default DockerMode: got %q, want %q", cfg.DockerMode, DockerModeCLI)
	}
}

func TestLoad_WithDBPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open() error: %v", err)
	}
	database.SetConfig("docker_mode", "api")
	database.Close()

	cfg, err := Load(dbPath)
	if err != nil {
		t.Errorf("Load() error: %v", err)
	}
	if cfg.DockerMode != DockerModeAPI {
		t.Errorf("Load() DockerMode: got %q, want %q", cfg.DockerMode, DockerModeAPI)
	}
}
