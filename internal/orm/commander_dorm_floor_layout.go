package orm

import (
	"context"
	"encoding/json"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderDormFloorLayout struct {
	CommanderID      uint32          `gorm:"primaryKey"`
	Floor            uint32          `gorm:"primaryKey"`
	FurniturePutList json.RawMessage `gorm:"type:json;not_null"`
}

func UpsertCommanderDormFloorLayoutTx(q *gen.Queries, commanderID uint32, floor uint32, furniturePutList json.RawMessage) error {
	ctx := context.Background()
	return q.UpsertCommanderDormFloorLayout(ctx, gen.UpsertCommanderDormFloorLayoutParams{
		CommanderID:      int64(commanderID),
		Floor:            int64(floor),
		FurniturePutList: []byte(furniturePutList),
	})
}

func ListCommanderDormFloorLayouts(commanderID uint32) ([]CommanderDormFloorLayout, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderDormFloorLayouts(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	entries := make([]CommanderDormFloorLayout, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, CommanderDormFloorLayout{
			CommanderID:      uint32(r.CommanderID),
			Floor:            uint32(r.Floor),
			FurniturePutList: json.RawMessage(r.FurniturePutList),
		})
	}
	return entries, nil
}
