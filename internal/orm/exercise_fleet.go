package orm

import (
	"context"
	"encoding/json"

	"github.com/ggmolly/belfast/internal/db"
)

type ExerciseFleet struct {
	CommanderID     uint32    `gorm:"primary_key"`
	VanguardShipIDs Int64List `gorm:"column=vanguard_ship_ids;type:json;not_null"`
	MainShipIDs     Int64List `gorm:"column=main_ship_ids;type:json;not_null"`
}

func GetExerciseFleet(commanderID uint32) (*ExerciseFleet, error) {
	ctx := context.Background()
	row := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id, vanguard_ship_ids, main_ship_ids
FROM exercise_fleets
WHERE commander_id = $1
`, int64(commanderID))
	var id int64
	var vanguardRaw []byte
	var mainRaw []byte
	err := row.Scan(&id, &vanguardRaw, &mainRaw)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	fleet := ExerciseFleet{CommanderID: uint32(id)}
	if len(vanguardRaw) > 0 {
		if err := json.Unmarshal(vanguardRaw, &fleet.VanguardShipIDs); err != nil {
			return nil, err
		}
	}
	if len(mainRaw) > 0 {
		if err := json.Unmarshal(mainRaw, &fleet.MainShipIDs); err != nil {
			return nil, err
		}
	}
	if fleet.VanguardShipIDs == nil {
		fleet.VanguardShipIDs = Int64List{}
	}
	if fleet.MainShipIDs == nil {
		fleet.MainShipIDs = Int64List{}
	}
	return &fleet, nil
}

func UpsertExerciseFleet(commanderID uint32, vanguardShipIDs []uint32, mainShipIDs []uint32) error {
	ctx := context.Background()
	vanguardRaw, err := json.Marshal(ToInt64List(vanguardShipIDs))
	if err != nil {
		return err
	}
	mainRaw, err := json.Marshal(ToInt64List(mainShipIDs))
	if err != nil {
		return err
	}
	_, err = db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO exercise_fleets (commander_id, vanguard_ship_ids, main_ship_ids)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id)
DO UPDATE SET
  vanguard_ship_ids = EXCLUDED.vanguard_ship_ids,
  main_ship_ids = EXCLUDED.main_ship_ids
`, int64(commanderID), vanguardRaw, mainRaw)
	return err
}
