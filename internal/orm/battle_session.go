package orm

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ggmolly/belfast/internal/db"
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

func GetBattleSession(commanderID uint32) (*BattleSession, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, system, stage_id, key, ship_ids, created_at, updated_at
FROM battle_sessions
WHERE commander_id = $1
`, int64(commanderID))
	var session BattleSession
	var shipIDsRaw []byte
	err := row.Scan(&session.CommanderID, &session.System, &session.StageID, &session.Key, &shipIDsRaw, &session.CreatedAt, &session.UpdatedAt)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	if len(shipIDsRaw) > 0 {
		if err := json.Unmarshal(shipIDsRaw, &session.ShipIDs); err != nil {
			return nil, err
		}
	}
	return &session, nil
}

func UpsertBattleSession(session *BattleSession) error {
	ctx := context.Background()
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now
	shipIDsRaw, err := json.Marshal([]int64(session.ShipIDs))
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO battle_sessions (
  commander_id,
  system,
  stage_id,
  key,
  ship_ids,
  created_at,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (commander_id)
DO UPDATE SET
  system = EXCLUDED.system,
  stage_id = EXCLUDED.stage_id,
  key = EXCLUDED.key,
  ship_ids = EXCLUDED.ship_ids,
  updated_at = EXCLUDED.updated_at
`, int64(session.CommanderID), int64(session.System), int64(session.StageID), int64(session.Key), shipIDsRaw, session.CreatedAt, session.UpdatedAt)
	return err
}

func DeleteBattleSession(commanderID uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `DELETE FROM battle_sessions WHERE commander_id = $1`, int64(commanderID))
	return err
}
