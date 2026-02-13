package orm

import (
	"context"
	"encoding/json"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderDormTheme struct {
	CommanderID      uint32          `gorm:"primaryKey"`
	ThemeSlotID      uint32          `gorm:"primaryKey"`
	Name             string          `gorm:"size:50;default:'';not_null"`
	FurniturePutList json.RawMessage `gorm:"type:json;not_null"`
}

func UpsertCommanderDormThemeTx(q *gen.Queries, commanderID uint32, slotID uint32, name string, furniturePutList json.RawMessage) error {
	ctx := context.Background()
	return q.UpsertCommanderDormTheme(ctx, gen.UpsertCommanderDormThemeParams{
		CommanderID:      int64(commanderID),
		ThemeSlotID:      int64(slotID),
		Name:             name,
		FurniturePutList: []byte(furniturePutList),
	})
}

func DeleteCommanderDormThemeTx(q *gen.Queries, commanderID uint32, slotID uint32) error {
	ctx := context.Background()
	return q.DeleteCommanderDormTheme(ctx, gen.DeleteCommanderDormThemeParams{CommanderID: int64(commanderID), ThemeSlotID: int64(slotID)})
}

func ListCommanderDormThemes(commanderID uint32) ([]CommanderDormTheme, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderDormThemes(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	entries := make([]CommanderDormTheme, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, CommanderDormTheme{
			CommanderID:      uint32(r.CommanderID),
			ThemeSlotID:      uint32(r.ThemeSlotID),
			Name:             r.Name,
			FurniturePutList: json.RawMessage(r.FurniturePutList),
		})
	}
	return entries, nil
}

func GetCommanderDormTheme(commanderID uint32, slotID uint32) (*CommanderDormTheme, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetCommanderDormTheme(ctx, gen.GetCommanderDormThemeParams{CommanderID: int64(commanderID), ThemeSlotID: int64(slotID)})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	entry := CommanderDormTheme{
		CommanderID:      uint32(row.CommanderID),
		ThemeSlotID:      uint32(row.ThemeSlotID),
		Name:             row.Name,
		FurniturePutList: json.RawMessage(row.FurniturePutList),
	}
	return &entry, nil
}
