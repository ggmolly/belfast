package orm

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SecondaryPasswordSettings struct {
	CommanderID  uint32    `gorm:"primaryKey;autoIncrement:false"`
	PasswordHash string    `gorm:"size:255;default:'';not_null"`
	Notice       string    `gorm:"size:200;default:'';not_null"`
	SystemList   Int64List `gorm:"type:json;not_null;default:'[]'"`
	FailCount    uint32    `gorm:"default:0;not_null"`
	FailCd       *int64    `gorm:"type:integer"`
}

func GetSecondaryPasswordSettings(db *gorm.DB, commanderID uint32) (SecondaryPasswordSettings, error) {
	var settings SecondaryPasswordSettings
	err := db.Where("commander_id = ?", commanderID).First(&settings).Error
	if err == nil {
		return settings, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return SecondaryPasswordSettings{
			CommanderID:  commanderID,
			PasswordHash: "",
			Notice:       "",
			SystemList:   Int64List{},
			FailCount:    0,
			FailCd:       nil,
		}, nil
	}
	return SecondaryPasswordSettings{}, err
}

func UpsertSecondaryPasswordSettings(db *gorm.DB, settings SecondaryPasswordSettings) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"password_hash", "notice", "system_list", "fail_count", "fail_cd"}),
	}).Create(&settings).Error
}

func UpdateSecondaryPasswordLockout(db *gorm.DB, commanderID uint32, failCount uint32, failCd *int64) error {
	return db.Model(&SecondaryPasswordSettings{}).
		Where("commander_id = ?", commanderID).
		Updates(map[string]any{
			"fail_count": failCount,
			"fail_cd":    failCd,
		}).Error
}

func ResetSecondaryPasswordLockout(db *gorm.DB, commanderID uint32) error {
	return UpdateSecondaryPasswordLockout(db, commanderID, 0, nil)
}
