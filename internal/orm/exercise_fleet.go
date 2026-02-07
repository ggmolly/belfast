package orm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExerciseFleet struct {
	CommanderID     uint32    `gorm:"primary_key"`
	VanguardShipIDs Int64List `gorm:"column=vanguard_ship_ids;type:json;not_null"`
	MainShipIDs     Int64List `gorm:"column=main_ship_ids;type:json;not_null"`
}

func GetExerciseFleet(db *gorm.DB, commanderID uint32) (*ExerciseFleet, error) {
	var fleet ExerciseFleet
	if err := db.Where("commander_id = ?", commanderID).First(&fleet).Error; err != nil {
		return nil, err
	}
	return &fleet, nil
}

func UpsertExerciseFleet(db *gorm.DB, commanderID uint32, vanguardShipIDs []uint32, mainShipIDs []uint32) error {
	fleet := ExerciseFleet{
		CommanderID:     commanderID,
		VanguardShipIDs: ToInt64List(vanguardShipIDs),
		MainShipIDs:     ToInt64List(mainShipIDs),
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"vanguard_ship_ids", "main_ship_ids"}),
	}).Create(&fleet).Error
}
