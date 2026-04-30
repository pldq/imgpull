package db

import (
	"testing"
)

func TestOpen_Success(t *testing.T) {
	// Use real sqlite for this test since glebarez/sqlite is pure Go
	db, err := Open(":memory:")
	if err != nil {
		t.Errorf("Open() error: %v", err)
	}
	if db == nil {
		t.Error("Open() returned nil")
	}
	defer db.Close()
}

func TestDB_Close(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
}
