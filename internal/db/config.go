package db

import (
	"errors"

	"gorm.io/gorm"
)

// ConfigItem represents a key-value config entry
type ConfigItem struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}

// GetConfig retrieves a config value by key
func (d *DB) GetConfig(key string) (string, error) {
	var item ConfigItem
	err := d.db.Where("key = ?", key).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return item.Value, nil
}

// SetConfig sets a config value
func (d *DB) SetConfig(key, value string) error {
	return d.db.Save(&ConfigItem{Key: key, Value: value}).Error
}
