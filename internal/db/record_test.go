package db

import (
	"testing"
	"time"
)

func TestImageRecord_UpsertRecord_Insert(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	now := time.Now()
	record := &ImageRecord{
		ImageName: "nginx",
		Tag:       "latest",
		Size:      1024,
		PullTime:  now,
		Duration:  5000,
	}

	if err := db.UpsertRecord(record); err != nil {
		t.Errorf("UpsertRecord() error: %v", err)
	}
	if record.ID == 0 {
		t.Error("UpsertRecord() should set ID")
	}
	// Compare within 1 second tolerance due to database time precision
	if record.FirstPullTime.Sub(now).Abs() > time.Second {
		t.Errorf("FirstPullTime should be set to PullTime on first insert")
	}
	if record.FirstDuration != 5000 {
		t.Errorf("FirstDuration should be set to Duration on first insert")
	}
}

func TestImageRecord_UpsertRecord_Update(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	now := time.Now()
	firstPullTime := now.Add(-2 * time.Hour)
	record1 := &ImageRecord{
		ImageName: "nginx",
		Tag:       "latest",
		Size:      1024,
		PullTime:  firstPullTime,
		Duration:  5000,
	}
	if err := db.UpsertRecord(record1); err != nil {
		t.Fatalf("UpsertRecord() first insert error: %v", err)
	}

	// Fetch the saved record to get actual FirstPullTime
	var saved ImageRecord
	db.db.First(&saved, record1.ID)
	originalFirstPullTime := saved.FirstPullTime

	record2 := &ImageRecord{
		ImageName: "nginx",
		Tag:       "latest",
		Size:      2048,
		PullTime:  now,
		Duration:  3000,
	}
	if err := db.UpsertRecord(record2); err != nil {
		t.Errorf("UpsertRecord() update error: %v", err)
	}

	records, _ := db.ListRecords(0)
	if len(records) != 1 {
		t.Errorf("Should have 1 record after update, got %d", len(records))
	}
	if records[0].Size != 2048 {
		t.Errorf("Size should be updated to 2048, got %d", records[0].Size)
	}
	if records[0].Duration != 3000 {
		t.Errorf("Duration should be updated to 3000, got %d", records[0].Duration)
	}
	if records[0].FirstDuration != 5000 {
		t.Errorf("FirstDuration should remain unchanged at 5000, got %d", records[0].FirstDuration)
	}
	if records[0].FirstPullTime.Sub(originalFirstPullTime).Abs() > time.Second {
		t.Errorf("FirstPullTime should remain unchanged")
	}
	if records[0].ID != record1.ID {
		t.Errorf("ID should remain same after update")
	}
}

func TestImageRecord_ListRecords(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.UpsertRecord(&ImageRecord{ImageName: "nginx", Tag: "latest", PullTime: now.Add(-1 * time.Hour), Duration: 5000})
	db.UpsertRecord(&ImageRecord{ImageName: "redis", Tag: "alpine", PullTime: now, Duration: 3000})

	records, err := db.ListRecords(0)
	if err != nil {
		t.Errorf("ListRecords() error: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("ListRecords() should return 2 records, got %d", len(records))
	}
	if records[0].ImageName != "redis" {
		t.Errorf("First record should be redis (latest pull_time), got %s", records[0].ImageName)
	}

	records, err = db.ListRecords(1)
	if err != nil {
		t.Errorf("ListRecords(1) error: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("ListRecords(1) should return 1 record, got %d", len(records))
	}
}
