package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
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

func GetOrCreateCommanderDormStateTx(ctx context.Context, q *gen.Queries, commanderID uint32) (*CommanderDormState, error) {
	row, err := q.GetCommanderDormStateByCommanderID(ctx, int64(commanderID))
	err = db.MapNotFound(err)
	if err == nil {
		state := CommanderDormState{
			CommanderID:            uint32(row.CommanderID),
			Level:                  uint32(row.Level),
			Food:                   uint32(row.Food),
			FoodMaxIncreaseCount:   uint32(row.FoodMaxIncreaseCount),
			FoodMaxIncrease:        uint32(row.FoodMaxIncrease),
			FloorNum:               uint32(row.FloorNum),
			ExpPos:                 uint32(row.ExpPos),
			NextTimestamp:          uint32(row.NextTimestamp),
			LoadExp:                uint32(row.LoadExp),
			LoadFood:               uint32(row.LoadFood),
			LoadTime:               uint32(row.LoadTime),
			UpdatedAtUnixTimestamp: uint32(row.UpdatedAtUnixTimestamp),
		}
		return &state, nil
	}
	if !db.IsNotFound(err) {
		return nil, err
	}

	state := CommanderDormState{CommanderID: commanderID}
	if state.Level == 0 {
		state.Level = 1
	}
	state.FloorNum = dormFloorNumForLevel(state.Level)
	state.UpdatedAtUnixTimestamp = uint32(time.Now().Unix())

	// Provide a sane initial food value from config if available.
	if entry, err := q.GetConfigEntry(ctx, gen.GetConfigEntryParams{Category: "ShareCfg/dorm_data_template.json", Key: fmt.Sprintf("%d", state.Level)}); err == nil {
		var tpl dormTemplate
		if err := json.Unmarshal(entry.Data, &tpl); err == nil && tpl.Capacity > 0 {
			state.Food = tpl.Capacity
		}
	}

	if err := q.CreateCommanderDormState(ctx, gen.CreateCommanderDormStateParams{
		CommanderID:            int64(state.CommanderID),
		Level:                  int64(state.Level),
		Food:                   int64(state.Food),
		FoodMaxIncreaseCount:   int64(state.FoodMaxIncreaseCount),
		FoodMaxIncrease:        int64(state.FoodMaxIncrease),
		FloorNum:               int64(state.FloorNum),
		ExpPos:                 int64(state.ExpPos),
		NextTimestamp:          int64(state.NextTimestamp),
		LoadExp:                int64(state.LoadExp),
		LoadFood:               int64(state.LoadFood),
		LoadTime:               int64(state.LoadTime),
		UpdatedAtUnixTimestamp: int64(state.UpdatedAtUnixTimestamp),
	}); err != nil {
		return nil, err
	}
	return &state, nil
}

func GetOrCreateCommanderDormState(commanderID uint32) (*CommanderDormState, error) {
	ctx := context.Background()
	var state *CommanderDormState
	err := db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		loaded, err := GetOrCreateCommanderDormStateTx(ctx, q, commanderID)
		if err != nil {
			return err
		}
		state = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

func SaveCommanderDormState(state *CommanderDormState) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commander_dorm_states
SET
  level = $2,
  food = $3,
  food_max_increase_count = $4,
  food_max_increase = $5,
  floor_num = $6,
  exp_pos = $7,
  next_timestamp = $8,
  load_exp = $9,
  load_food = $10,
  load_time = $11,
  updated_at_unix_timestamp = $12
WHERE commander_id = $1
`, int64(state.CommanderID), int64(state.Level), int64(state.Food), int64(state.FoodMaxIncreaseCount), int64(state.FoodMaxIncrease), int64(state.FloorNum), int64(state.ExpPos), int64(state.NextTimestamp), int64(state.LoadExp), int64(state.LoadFood), int64(state.LoadTime), int64(state.UpdatedAtUnixTimestamp))
	return err
}

func SaveCommanderDormStateTx(ctx context.Context, tx pgx.Tx, state *CommanderDormState) error {
	_, err := tx.Exec(ctx, `
UPDATE commander_dorm_states
SET
  level = $2,
  food = $3,
  food_max_increase_count = $4,
  food_max_increase = $5,
  floor_num = $6,
  exp_pos = $7,
  next_timestamp = $8,
  load_exp = $9,
  load_food = $10,
  load_time = $11,
  updated_at_unix_timestamp = $12
WHERE commander_id = $1
`, int64(state.CommanderID), int64(state.Level), int64(state.Food), int64(state.FoodMaxIncreaseCount), int64(state.FoodMaxIncrease), int64(state.FloorNum), int64(state.ExpPos), int64(state.NextTimestamp), int64(state.LoadExp), int64(state.LoadFood), int64(state.LoadTime), int64(state.UpdatedAtUnixTimestamp))
	return err
}
