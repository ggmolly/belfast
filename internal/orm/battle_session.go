package orm

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BattleSession struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	System      uint32    `gorm:"not_null"`
	StageID     uint32    `gorm:"not_null"`
	Key         uint32    `gorm:"not_null"`
	ShipIDs     Int64List `gorm:"type:text;not_null"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func GetBattleSession(db *gorm.DB, commanderID uint32) (*BattleSession, error) {
	var session BattleSession
	if err := db.Where("commander_id = ?", commanderID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func UpsertBattleSession(db *gorm.DB, session *BattleSession) error {
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"system", "stage_id", "key", "ship_ids", "updated_at"}),
	}).Create(session).Error
}

func DeleteBattleSession(db *gorm.DB, commanderID uint32) error {
	return db.Where("commander_id = ?", commanderID).Delete(&BattleSession{}).Error
}
