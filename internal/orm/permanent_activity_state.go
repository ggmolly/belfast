package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type PermanentActivityState struct {
	CommanderID         uint32    `gorm:"primaryKey;autoIncrement:false"`
	PermanentNow        uint32    `gorm:"not_null;default:0"`
	FinishedActivityIDs Int64List `gorm:"type:json;not_null;default:'[]'"`
	CreatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt           time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func GetOrCreatePermanentActivityState(db *gorm.DB, commanderID uint32) (*PermanentActivityState, error) {
	var state PermanentActivityState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = PermanentActivityState{
			CommanderID:         commanderID,
			PermanentNow:        0,
			FinishedActivityIDs: Int64List{},
		}
		if err := db.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	if state.FinishedActivityIDs == nil {
		state.FinishedActivityIDs = Int64List{}
	}
	return &state, nil
}
