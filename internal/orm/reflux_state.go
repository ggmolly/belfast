package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
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

func GetOrCreateRefluxState(commanderID uint32) (*RefluxState, error) {
	ctx := context.Background()
	state := RefluxState{}
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, active, return_lv, return_time, ship_number, last_offline_time, pt, sign_cnt, sign_last_time, pt_stage, created_at, updated_at
FROM reflux_states
WHERE commander_id = $1
`, int64(commanderID)).Scan(
		&state.CommanderID,
		&state.Active,
		&state.ReturnLv,
		&state.ReturnTime,
		&state.ShipNumber,
		&state.LastOfflineTime,
		&state.Pt,
		&state.SignCnt,
		&state.SignLastTime,
		&state.PtStage,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	err = db.MapNotFound(err)
	if err == nil {
		return &state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}
	state = RefluxState{CommanderID: commanderID}
	if err := SaveRefluxState(&state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveRefluxState(state *RefluxState) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO reflux_states (commander_id, active, return_lv, return_time, ship_number, last_offline_time, pt, sign_cnt, sign_last_time, pt_stage, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
ON CONFLICT (commander_id)
DO UPDATE SET
  active = EXCLUDED.active,
  return_lv = EXCLUDED.return_lv,
  return_time = EXCLUDED.return_time,
  ship_number = EXCLUDED.ship_number,
  last_offline_time = EXCLUDED.last_offline_time,
  pt = EXCLUDED.pt,
  sign_cnt = EXCLUDED.sign_cnt,
  sign_last_time = EXCLUDED.sign_last_time,
  pt_stage = EXCLUDED.pt_stage,
  updated_at = NOW()
`, int64(state.CommanderID), int64(state.Active), int64(state.ReturnLv), int64(state.ReturnTime), int64(state.ShipNumber), int64(state.LastOfflineTime), int64(state.Pt), int64(state.SignCnt), int64(state.SignLastTime), int64(state.PtStage))
	return err
}
