package orm

import (
	"errors"

	"gorm.io/gorm"
)

type SecondaryPasswordState struct {
	CommanderID  uint32    `gorm:"primaryKey;autoIncrement:false"`
	PasswordHash string    `gorm:"type:text;not_null;default:''"`
	Notice       string    `gorm:"type:text;not_null;default:''"`
	SystemList   Int64List `gorm:"type:text;not_null;default:'[]'"`
	State        uint32    `gorm:"not_null;default:0"`
	FailCount    uint32    `gorm:"not_null;default:0"`
	FailCd       uint32    `gorm:"not_null;default:0"`
}

func GetOrCreateSecondaryPasswordState(db *gorm.DB, commanderID uint32) (*SecondaryPasswordState, error) {
	var state SecondaryPasswordState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = SecondaryPasswordState{
			CommanderID:  commanderID,
			PasswordHash: "",
			Notice:       "",
			SystemList:   Int64List{},
			State:        0,
			FailCount:    0,
			FailCd:       0,
		}
		if err := db.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	if state.SystemList == nil {
		state.SystemList = Int64List{}
	}
	return &state, nil
}

func SaveSecondaryPasswordState(db *gorm.DB, state *SecondaryPasswordState) error {
	return db.Save(state).Error
}
