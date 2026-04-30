package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ImageRecord represents a pulled image record
type ImageRecord struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	ImageName     string    `gorm:"not null;uniqueIndex:idx_image_tag"`
	Tag           string    `gorm:"not null;uniqueIndex:idx_image_tag"`
	Size          int64     `gorm:"default:0"`
	PullTime      time.Time `gorm:"not null;index"`
	Duration      int64     `gorm:"default:0"` // Duration in milliseconds
	FirstPullTime time.Time `gorm:"not null"`  // First pull time (never changes)
	FirstDuration int64     `gorm:"default:0"` // First pull duration in milliseconds
}

// UpsertRecord inserts or updates a record based on image_name + tag
func (d *DB) UpsertRecord(record *ImageRecord) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		var existing ImageRecord
		err := tx.Where("image_name = ? AND tag = ?", record.ImageName, record.Tag).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// First pull: set FirstPullTime and FirstDuration
			record.FirstPullTime = record.PullTime
			record.FirstDuration = record.Duration
			return tx.Create(record).Error
		}
		if err != nil {
			return err
		}
		// Update: keep FirstPullTime and FirstDuration, update others
		record.ID = existing.ID
		record.FirstPullTime = existing.FirstPullTime
		record.FirstDuration = existing.FirstDuration
		return tx.Save(record).Error
	})
}

// ListRecords returns all image records
func (d *DB) ListRecords(limit int) ([]ImageRecord, error) {
	var records []ImageRecord
	query := d.db.Order("pull_time desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&records).Error
	return records, err
}
