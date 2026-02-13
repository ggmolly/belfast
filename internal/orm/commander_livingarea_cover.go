package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderLivingAreaCover struct {
	CommanderID uint32    `gorm:"primaryKey;autoIncrement:false"`
	CoverID     uint32    `gorm:"primaryKey;autoIncrement:false"`
	UnlockedAt  time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;not_null"`
	IsNew       bool      `gorm:"default:false;not_null"`
}

func ListCommanderLivingAreaCovers(commanderID uint32) ([]CommanderLivingAreaCover, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderLivingAreaCovers(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	entries := make([]CommanderLivingAreaCover, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, CommanderLivingAreaCover{
			CommanderID: uint32(r.CommanderID),
			CoverID:     uint32(r.CoverID),
			UnlockedAt:  r.UnlockedAt.Time,
			IsNew:       r.IsNew,
		})
	}
	return entries, nil
}

func CommanderHasLivingAreaCover(commanderID uint32, coverID uint32) (bool, error) {
	ctx := context.Background()
	_, err := db.DefaultStore.Queries.GetCommanderLivingAreaCover(ctx, gen.GetCommanderLivingAreaCoverParams{CommanderID: int64(commanderID), CoverID: int64(coverID)})
	err = db.MapNotFound(err)
	if db.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func UpsertCommanderLivingAreaCover(entry CommanderLivingAreaCover) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertCommanderLivingAreaCover(ctx, gen.UpsertCommanderLivingAreaCoverParams{
		CommanderID: int64(entry.CommanderID),
		CoverID:     int64(entry.CoverID),
		UnlockedAt:  pgTimestamptz(entry.UnlockedAt),
		IsNew:       entry.IsNew,
	})
}

func GetCommanderLivingAreaCoverEntry(commanderID uint32, coverID uint32) (*CommanderLivingAreaCover, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetCommanderLivingAreaCover(ctx, gen.GetCommanderLivingAreaCoverParams{
		CommanderID: int64(commanderID),
		CoverID:     int64(coverID),
	})
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	entry := CommanderLivingAreaCover{
		CommanderID: uint32(row.CommanderID),
		CoverID:     uint32(row.CoverID),
		UnlockedAt:  row.UnlockedAt.Time,
		IsNew:       row.IsNew,
	}
	return &entry, nil
}

func UpdateCommanderLivingAreaCoverIsNew(commanderID uint32, coverID uint32, isNew bool) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commander_living_area_covers
SET is_new = $3
WHERE commander_id = $1
  AND cover_id = $2
`, int64(commanderID), int64(coverID), isNew)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteCommanderLivingAreaCover(commanderID uint32, coverID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_living_area_covers
WHERE commander_id = $1
  AND cover_id = $2
`, int64(commanderID), int64(coverID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
