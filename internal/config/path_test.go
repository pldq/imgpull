package config

import (
	"path/filepath"
	"testing"
)

func TestGetDBPath(t *testing.T) {
	path, err := GetDBPath()
	if err != nil {
		t.Errorf("GetDBPath() error: %v", err)
	}
	if path == "" {
		t.Error("GetDBPath() returned empty path")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("GetDBPath() should return absolute path, got %q", path)
	}
}

func TestGetConfigDir(t *testing.T) {
	dir, err := getConfigDir()
	if err != nil {
		t.Errorf("getConfigDir() error: %v", err)
	}
	if dir == "" {
		t.Error("getConfigDir() returned empty dir")
	}
	if filepath.Base(dir) != appDirName {
		t.Errorf("getConfigDir() should end with %q, got %q", appDirName, filepath.Base(dir))
	}
}
