package answer

// TODO(dorm-sim): This is a minimal dorm ticking implementation.
// It intentionally approximates client behavior using ShareCfg/dorm_data_template.json.
// Remaining parity work:
// - Confirm exact food/exp/comfort formulas and per-ship caps.
// - Confirm when exp is granted (continuous vs on-exit) and how SC_19009/SC_19010 are used.
// - Confirm NextTimestamp semantics across empty dorm / out-of-food states.
// - Confirm intimacy and dorm_icon accumulation rates (and any modifiers).
// - Confirm per-floor map sizing and rules for multi-floor dorms.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"
)

type dormLevelTemplate struct {
	ID          uint32 `json:"id"`
	Capacity    uint32 `json:"capacity"`
	Consume     uint32 `json:"consume"`
	Exp         uint32 `json:"exp"`
	Time        uint32 `json:"time"`
	Comfortable uint32 `json:"comfortable"`
}

func loadDormLevelTemplate(level uint32) (*dormLevelTemplate, error) {
	if level == 0 {
		level = 1
	}
	entry, err := orm.GetConfigEntry("ShareCfg/dorm_data_template.json", fmt.Sprintf("%d", level))
	if err != nil {
		return nil, err
	}
	var tpl dormLevelTemplate
	if err := json.Unmarshal(entry.Data, &tpl); err != nil {
		return nil, err
	}
	if tpl.ID == 0 {
		tpl.ID = level
	}
	return &tpl, nil
}

type dormTickResult struct {
	PopList     []*protobuf.POP_INFO
	ExpGained   uint32
	FoodConsume uint32
}

func tickDormStateTx(ctx context.Context, tx pgx.Tx, commanderID uint32, now uint32) (*dormTickResult, error) {
	q := db.DefaultStore.Queries.WithTx(tx)
	state, err := orm.GetOrCreateCommanderDormStateTx(ctx, q, commanderID)
	if err != nil {
		return nil, err
	}

	tpl, err := loadDormLevelTemplate(state.Level)
	if err != nil {
		// Config missing => no simulation.
		return &dormTickResult{}, nil
	}
	if tpl.Time == 0 {
		return &dormTickResult{}, nil
	}

	// Ships currently in dorm (train/rest).
	rows, err := tx.Query(ctx, `
SELECT id, ship_id, level, exp, surplus_exp, max_level, intimacy, state, state_info1, state_info2, state_info3, state_info4
FROM owned_ships
WHERE owner_id = $1
  AND deleted_at IS NULL
  AND state IN (5, 2)
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dormShips := make([]orm.OwnedShip, 0)
	for rows.Next() {
		var ship orm.OwnedShip
		var id int64
		var shipID int64
		if err := rows.Scan(&id, &shipID, &ship.Level, &ship.Exp, &ship.SurplusExp, &ship.MaxLevel, &ship.Intimacy, &ship.State, &ship.StateInfo1, &ship.StateInfo2, &ship.StateInfo3, &ship.StateInfo4); err != nil {
			return nil, err
		}
		ship.OwnerID = commanderID
		ship.ID = uint32(id)
		ship.ShipID = uint32(shipID)
		dormShips = append(dormShips, ship)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(dormShips) == 0 {
		state.NextTimestamp = 0
		state.UpdatedAtUnixTimestamp = now
		if err := orm.SaveCommanderDormStateTx(ctx, tx, state); err != nil {
			return nil, err
		}
		return &dormTickResult{}, nil
	}

	last := state.UpdatedAtUnixTimestamp
	if last == 0 {
		state.UpdatedAtUnixTimestamp = now
		if err := orm.SaveCommanderDormStateTx(ctx, tx, state); err != nil {
			return nil, err
		}
		return &dormTickResult{}, nil
	}
	if now <= last {
		return &dormTickResult{}, nil
	}

	elapsed := now - last
	ticks := elapsed / tpl.Time
	if ticks == 0 {
		return &dormTickResult{}, nil
	}

	shipCount := uint32(len(dormShips))
	perTickFood := tpl.Consume * shipCount
	if perTickFood == 0 {
		return &dormTickResult{}, nil
	}
	maxTicksByFood := state.Food / perTickFood
	if maxTicksByFood == 0 {
		state.NextTimestamp = 0
		state.UpdatedAtUnixTimestamp = now
		if err := orm.SaveCommanderDormStateTx(ctx, tx, state); err != nil {
			return nil, err
		}
		return &dormTickResult{}, nil
	}
	if ticks > maxTicksByFood {
		ticks = maxTicksByFood
	}

	coinPerTick := uint32(1)
	if tpl.Comfortable > 0 {
		coinPerTick += tpl.Comfortable / 10
	}
	intimacyPerTick := uint32(1)

	trainingCount := uint32(0)
	for i := range dormShips {
		if dormShips[i].State == 5 {
			trainingCount++
		}
	}

	foodConsume := perTickFood * ticks
	state.Food -= foodConsume
	state.LoadFood += foodConsume

	expGained := tpl.Exp * trainingCount * ticks
	state.LoadExp += expGained
	state.LoadTime = now
	state.UpdatedAtUnixTimestamp = last + ticks*tpl.Time
	if trainingCount > 0 && state.Food >= perTickFood {
		state.NextTimestamp = state.UpdatedAtUnixTimestamp + tpl.Time
	} else {
		state.NextTimestamp = 0
	}

	popList := make([]*protobuf.POP_INFO, 0, len(dormShips))
	for i := range dormShips {
		ship := &dormShips[i]
		ship.StateInfo3 += intimacyPerTick * ticks
		ship.StateInfo4 += coinPerTick * ticks
		if ship.State == 5 {
			ship.StateInfo2 += tpl.Exp * ticks
		}
		if err := saveDormOwnedShipTx(ctx, tx, ship); err != nil {
			return nil, err
		}
		popList = append(popList, &protobuf.POP_INFO{Id: proto.Uint32(ship.ID), Intimacy: proto.Uint32(intimacyPerTick * ticks), DormIcon: proto.Uint32(coinPerTick * ticks)})
	}
	if err := orm.SaveCommanderDormStateTx(ctx, tx, state); err != nil {
		return nil, err
	}

	return &dormTickResult{PopList: popList, ExpGained: expGained, FoodConsume: foodConsume}, nil
}

func tickDormAndPush(client *connection.Client) error {
	commanderID := client.Commander.CommanderID
	now := uint32(time.Now().Unix())

	ctx := context.Background()
	var res *dormTickResult
	err := orm.WithPGXTx(ctx, func(tx pgx.Tx) error {
		loaded, err := tickDormStateTx(ctx, tx, commanderID, now)
		if err != nil {
			return err
		}
		res = loaded
		return nil
	})
	if err != nil {
		return err
	}

	if res != nil && len(res.PopList) > 0 {
		_, _, err := client.SendMessage(19010, &protobuf.SC_19010{PopList: res.PopList})
		return err
	}
	return nil
}

func saveDormOwnedShipTx(ctx context.Context, tx pgx.Tx, ship *orm.OwnedShip) error {
	_, err := tx.Exec(ctx, `
UPDATE owned_ships
SET
  level = $3,
  exp = $4,
  surplus_exp = $5,
  intimacy = $6,
  state = $7,
  state_info1 = $8,
  state_info2 = $9,
  state_info3 = $10,
  state_info4 = $11
WHERE owner_id = $1
  AND id = $2
  AND deleted_at IS NULL
`, int64(ship.OwnerID), int64(ship.ID), int64(ship.Level), int64(ship.Exp), int64(ship.SurplusExp), int64(ship.Intimacy), int64(ship.State), int64(ship.StateInfo1), int64(ship.StateInfo2), int64(ship.StateInfo3), int64(ship.StateInfo4))
	return err
}
