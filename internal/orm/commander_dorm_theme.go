package orm

import (
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderDormTheme struct {
	CommanderID      uint32          `gorm:"primaryKey"`
	ThemeSlotID      uint32          `gorm:"primaryKey"`
	Name             string          `gorm:"size:50;default:'';not_null"`
	FurniturePutList json.RawMessage `gorm:"type:json;not_null"`
}

func UpsertCommanderDormThemeTx(tx *gorm.DB, commanderID uint32, slotID uint32, name string, furniturePutList json.RawMessage) error {
	entry := CommanderDormTheme{CommanderID: commanderID, ThemeSlotID: slotID, Name: name, FurniturePutList: furniturePutList}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "theme_slot_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"name":               name,
			"furniture_put_list": furniturePutList,
		}),
	}).Create(&entry).Error
}

func DeleteCommanderDormThemeTx(tx *gorm.DB, commanderID uint32, slotID uint32) error {
	return tx.Where("commander_id = ? AND theme_slot_id = ?", commanderID, slotID).Delete(&CommanderDormTheme{}).Error
}

func ListCommanderDormThemes(commanderID uint32) ([]CommanderDormTheme, error) {
	var entries []CommanderDormTheme
	if err := GormDB.Where("commander_id = ?", commanderID).Order("theme_slot_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func GetCommanderDormTheme(commanderID uint32, slotID uint32) (*CommanderDormTheme, error) {
	var entry CommanderDormTheme
	if err := GormDB.Where("commander_id = ? AND theme_slot_id = ?", commanderID, slotID).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}
