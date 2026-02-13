package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
	"github.com/jackc/pgx/v5"
)

type CommanderFurniture struct {
	CommanderID uint32 `gorm:"not_null;primaryKey"`
	FurnitureID uint32 `gorm:"not_null;primaryKey"`
	Count       uint32 `gorm:"not_null"`
	GetTime     uint32 `gorm:"not_null"`
}

func ListCommanderFurniture(commanderID uint32) ([]CommanderFurniture, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderFurnitures(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	entries := make([]CommanderFurniture, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, CommanderFurniture{
			CommanderID: uint32(r.CommanderID),
			FurnitureID: uint32(r.FurnitureID),
			Count:       uint32(r.Count),
			GetTime:     uint32(r.GetTime),
		})
	}
	return entries, nil
}

func AddCommanderFurnitureTx(ctx context.Context, tx pgx.Tx, commanderID uint32, furnitureID uint32, count uint32, getTime uint32) error {
	q := gen.New(tx)
	return q.AddCommanderFurniture(ctx, gen.AddCommanderFurnitureParams{
		CommanderID: int64(commanderID),
		FurnitureID: int64(furnitureID),
		Count:       int64(count),
		GetTime:     int64(getTime),
	})
}
