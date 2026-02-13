package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

// CommanderMedalDisplay persists a commander's ordered medal/trophy display list.
// Ordering is stable via Position (0-based).
type CommanderMedalDisplay struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	Position    uint32 `gorm:"primaryKey;autoIncrement:false"`
	MedalID     uint32 `gorm:"not null"`
}

func ListCommanderMedalDisplay(commanderID uint32) ([]uint32, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderMedalDisplay(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	medalIDs := make([]uint32, 0, len(rows))
	for _, r := range rows {
		medalIDs = append(medalIDs, uint32(r.MedalID))
	}
	return medalIDs, nil
}

func SetCommanderMedalDisplay(commanderID uint32, medalIDs []uint32) error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.DeleteCommanderMedalDisplayByCommanderID(ctx, int64(commanderID)); err != nil {
			return err
		}
		for i, medalID := range medalIDs {
			if err := q.CreateCommanderMedalDisplayRow(ctx, gen.CreateCommanderMedalDisplayRowParams{
				CommanderID: int64(commanderID),
				Position:    int64(i),
				MedalID:     int64(medalID),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}
