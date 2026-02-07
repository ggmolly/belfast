package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommanderFurniture struct {
	CommanderID uint32 `gorm:"not_null;primaryKey"`
	FurnitureID uint32 `gorm:"not_null;primaryKey"`
	Count       uint32 `gorm:"not_null"`
	GetTime     uint32 `gorm:"not_null"`
}

func ListCommanderFurniture(commanderID uint32) ([]CommanderFurniture, error) {
	var entries []CommanderFurniture
	if err := GormDB.Where("commander_id = ?", commanderID).Order("furniture_id asc").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func AddCommanderFurnitureTx(tx *gorm.DB, commanderID uint32, furnitureID uint32, count uint32, getTime uint32) error {
	entry := CommanderFurniture{CommanderID: commanderID, FurnitureID: furnitureID, Count: count, GetTime: getTime}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "commander_id"}, {Name: "furniture_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"count":    gorm.Expr("count + ?", count),
			"get_time": getTime,
		}),
	}).Create(&entry).Error
}
