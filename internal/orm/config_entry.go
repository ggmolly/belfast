package orm

import (
	"encoding/json"

	"gorm.io/gorm"
)

type ConfigEntry struct {
	ID       uint64          `gorm:"primary_key"`
	Category string          `gorm:"size:160;not_null;index:idx_config_category_key,unique"`
	Key      string          `gorm:"size:128;not_null;index:idx_config_category_key,unique"`
	Data     json.RawMessage `gorm:"type:json;not_null"`
}

func ListConfigEntries(db *gorm.DB, category string) ([]ConfigEntry, error) {
	var entries []ConfigEntry
	if err := db.Where("category = ?", category).Order("key asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func GetConfigEntry(db *gorm.DB, category string, key string) (*ConfigEntry, error) {
	var entry ConfigEntry
	if err := db.Where("category = ? AND key = ?", category, key).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}
