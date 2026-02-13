package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderAttire struct {
	CommanderID uint32     `gorm:"primaryKey;autoIncrement:false"`
	Type        uint32     `gorm:"primaryKey;autoIncrement:false"`
	AttireID    uint32     `gorm:"primaryKey;autoIncrement:false"`
	ExpiresAt   *time.Time `gorm:"type:timestamp"`
	IsNew       bool       `gorm:"default:false;not_null"`
}

func ListCommanderAttires(commanderID uint32) ([]CommanderAttire, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderAttires(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	entries := make([]CommanderAttire, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, CommanderAttire{
			CommanderID: uint32(r.CommanderID),
			Type:        uint32(r.Type),
			AttireID:    uint32(r.AttireID),
			ExpiresAt:   pgTimestamptzPtr(r.ExpiresAt),
			IsNew:       r.IsNew,
		})
	}
	return entries, nil
}

func ListCommanderAttiresByType(commanderID uint32, attireType uint32) ([]CommanderAttire, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderAttiresByType(ctx, gen.ListCommanderAttiresByTypeParams{CommanderID: int64(commanderID), Type: int64(attireType)})
	if err != nil {
		return nil, err
	}
	entries := make([]CommanderAttire, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, CommanderAttire{
			CommanderID: uint32(r.CommanderID),
			Type:        uint32(r.Type),
			AttireID:    uint32(r.AttireID),
			ExpiresAt:   pgTimestamptzPtr(r.ExpiresAt),
			IsNew:       r.IsNew,
		})
	}
	return entries, nil
}

func CommanderHasAttire(commanderID uint32, attireType uint32, attireID uint32, now time.Time) (bool, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetCommanderAttire(ctx, gen.GetCommanderAttireParams{CommanderID: int64(commanderID), Type: int64(attireType), AttireID: int64(attireID)})
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	expiresAt := pgTimestamptzPtr(row.ExpiresAt)
	if expiresAt != nil && expiresAt.Before(now) {
		return false, nil
	}
	return true, nil
}

func UpsertCommanderAttire(entry CommanderAttire) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertCommanderAttire(ctx, gen.UpsertCommanderAttireParams{
		CommanderID: int64(entry.CommanderID),
		Type:        int64(entry.Type),
		AttireID:    int64(entry.AttireID),
		ExpiresAt:   pgTimestamptzFromPtr(entry.ExpiresAt),
		IsNew:       entry.IsNew,
	})
}

func GetCommanderAttireEntry(commanderID uint32, attireType uint32, attireID uint32) (*CommanderAttire, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetCommanderAttire(ctx, gen.GetCommanderAttireParams{
		CommanderID: int64(commanderID),
		Type:        int64(attireType),
		AttireID:    int64(attireID),
	})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	entry := CommanderAttire{
		CommanderID: uint32(row.CommanderID),
		Type:        uint32(row.Type),
		AttireID:    uint32(row.AttireID),
		ExpiresAt:   pgTimestamptzPtr(row.ExpiresAt),
		IsNew:       row.IsNew,
	}
	return &entry, nil
}

func DeleteCommanderAttire(commanderID uint32, attireType uint32, attireID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_attires
WHERE commander_id = $1
  AND type = $2
  AND attire_id = $3
`, int64(commanderID), int64(attireType), int64(attireID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
