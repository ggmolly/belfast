package orm

import (
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderDormFloorLayout struct {
	CommanderID      uint32          `gorm:"primaryKey"`
	Floor            uint32          `gorm:"primaryKey"`
	FurniturePutList json.RawMessage `gorm:"type:json;not_null"`
}

func UpsertCommanderDormFloorLayoutTx(tx *gorm.DB, commanderID uint32, floor uint32, furniturePutList json.RawMessage) error {
	entry := CommanderDormFloorLayout{CommanderID: commanderID, Floor: floor, FurniturePutList: furniturePutList}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "floor"}},
		DoUpdates: clause.Assignments(map[string]any{"furniture_put_list": furniturePutList}),
	}).Create(&entry).Error
}

func ListCommanderDormFloorLayouts(commanderID uint32) ([]CommanderDormFloorLayout, error) {
	var entries []CommanderDormFloorLayout
	if err := GormDB.Where("commander_id = ?", commanderID).Order("floor asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
