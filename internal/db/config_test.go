package db

import (
	"testing"
)

func TestDB_GetConfig(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	value, err := db.GetConfig("nonexistent")
	if err != nil {
		t.Errorf("GetConfig() should not error for non-existent key: %v", err)
	}
	if value != "" {
		t.Errorf("GetConfig() should return empty for non-existent key, got %q", value)
	}

	if err := db.SetConfig("test_key", "test_value"); err != nil {
		t.Fatalf("SetConfig() error: %v", err)
	}

	value, err = db.GetConfig("test_key")
	if err != nil {
		t.Errorf("GetConfig() error: %v", err)
	}
	if value != "test_value" {
		t.Errorf("GetConfig() value: got %q, want %q", value, "test_value")
	}
}

func TestDB_SetConfig(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	if err := db.SetConfig("key1", "value1"); err != nil {
		t.Errorf("SetConfig() error: %v", err)
	}

	if err := db.SetConfig("key1", "value2"); err != nil {
		t.Errorf("SetConfig() update error: %v", err)
	}

	value, _ := db.GetConfig("key1")
	if value != "value2" {
		t.Errorf("SetConfig() should update value, got %q, want %q", value, "value2")
	}
}
