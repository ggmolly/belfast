package orm

import (
	"time"

	"gorm.io/gorm/clause"
)

type CommanderBuff struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	BuffID      uint32    `gorm:"primaryKey;autoIncrement:false"`
	ExpiresAt   time.Time `gorm:"not_null;index:idx_commander_buff_expires_at"`
}

func ListCommanderBuffs(commanderID uint32) ([]CommanderBuff, error) {
	var buffs []CommanderBuff
	if err := GormDB.
		Where("commander_id = ?", commanderID).
		Order("buff_id asc").
		Find(&buffs).
		Error; err != nil {
		return nil, err
	}
	return buffs, nil
}

func UpsertCommanderBuff(commanderID uint32, buffID uint32, expiresAt time.Time) error {
	entry := CommanderBuff{
		CommanderID: commanderID,
		BuffID:      buffID,
		ExpiresAt:   expiresAt.UTC(),
	}
	return GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "buff_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"expires_at"}),
	}).Create(&entry).Error
}

func ListCommanderActiveBuffs(commanderID uint32, now time.Time) ([]CommanderBuff, error) {
	var buffs []CommanderBuff
	if err := GormDB.
		Where("commander_id = ? AND expires_at > ?", commanderID, now).
		Find(&buffs).
		Error; err != nil {
		return nil, err
	}
	return buffs, nil
}
