package orm

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type dormTemplate struct {
	ID       uint32 `json:"id"`
	Capacity uint32 `json:"capacity"`
}

// CommanderDormState stores persistent backyard/dorm state surfaced in SC_19001.
// Keep this minimal; deltas (SC_19009/SC_19010) are handled separately.
type CommanderDormState struct {
	CommanderID            uint32 `gorm:"primaryKey"`
	Level                  uint32 `gorm:"not_null;default:1"`
	Food                   uint32 `gorm:"not_null;default:0"`
	FoodMaxIncreaseCount   uint32 `gorm:"not_null;default:0"`
	FoodMaxIncrease        uint32 `gorm:"not_null;default:0"`
	FloorNum               uint32 `gorm:"not_null;default:1"`
	ExpPos                 uint32 `gorm:"not_null;default:2"`
	NextTimestamp          uint32 `gorm:"not_null;default:0"`
	LoadExp                uint32 `gorm:"not_null;default:0"`
	LoadFood               uint32 `gorm:"not_null;default:0"`
	LoadTime               uint32 `gorm:"not_null;default:0"`
	UpdatedAtUnixTimestamp uint32 `gorm:"not_null;default:0"`
}

func dormFloorNumForLevel(level uint32) uint32 {
	if level == 0 {
		return 1
	}
	if level >= 3 {
		return 3
	}
	return level
}

func GetOrCreateCommanderDormStateTx(tx *gorm.DB, commanderID uint32) (*CommanderDormState, error) {
	var state CommanderDormState
	if err := tx.Where("commander_id = ?", commanderID).First(&state).Error; err == nil {
		return &state, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	state = CommanderDormState{CommanderID: commanderID}
	if state.Level == 0 {
		state.Level = 1
	}
	state.FloorNum = dormFloorNumForLevel(state.Level)
	state.UpdatedAtUnixTimestamp = uint32(time.Now().Unix())
	// Provide a sane initial food value from config if available.
	if entry, err := GetConfigEntry(tx, "ShareCfg/dorm_data_template.json", fmt.Sprintf("%d", state.Level)); err == nil {
		var tpl dormTemplate
		if err := json.Unmarshal(entry.Data, &tpl); err == nil && tpl.Capacity > 0 {
			state.Food = tpl.Capacity
		}
	}
	if err := tx.Create(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func GetOrCreateCommanderDormState(commanderID uint32) (*CommanderDormState, error) {
	return GetOrCreateCommanderDormStateTx(GormDB, commanderID)
}
