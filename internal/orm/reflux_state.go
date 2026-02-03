package orm

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type RefluxState struct {
	CommanderID     uint32    `gorm:"primaryKey;autoIncrement:false"`
	Active          uint32    `gorm:"not_null;default:0"`
	ReturnLv        uint32    `gorm:"not_null;default:0"`
	ReturnTime      uint32    `gorm:"not_null;default:0"`
	ShipNumber      uint32    `gorm:"not_null;default:0"`
	LastOfflineTime uint32    `gorm:"not_null;default:0"`
	Pt              uint32    `gorm:"not_null;default:0"`
	SignCnt         uint32    `gorm:"not_null;default:0"`
	SignLastTime    uint32    `gorm:"not_null;default:0"`
	PtStage         uint32    `gorm:"not_null;default:0"`
	CreatedAt       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	UpdatedAt       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
}

func GetOrCreateRefluxState(db *gorm.DB, commanderID uint32) (*RefluxState, error) {
	var state RefluxState
	if err := db.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		state = RefluxState{
			CommanderID:     commanderID,
			Active:          0,
			ReturnLv:        0,
			ReturnTime:      0,
			ShipNumber:      0,
			LastOfflineTime: 0,
			Pt:              0,
			SignCnt:         0,
			SignLastTime:    0,
			PtStage:         0,
		}
		if err := db.Create(&state).Error; err != nil {
			return nil, err
		}
	}
	return &state, nil
}

func SaveRefluxState(db *gorm.DB, state *RefluxState) error {
	return db.Save(state).Error
}
