package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type CommanderCommonFlag struct {
	CommanderID uint32
	FlagID      uint32
}

func ListCommanderCommonFlags(commanderID uint32) ([]uint32, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListCommanderCommonFlags(ctx, int64(commanderID))
	if err != nil {
		return nil, err
	}
	flags := make([]uint32, 0, len(rows))
	for _, id := range rows {
		flags = append(flags, uint32(id))
	}
	return flags, nil
}

func SetCommanderCommonFlag(commanderID uint32, flagID uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.CreateCommanderCommonFlag(ctx, gen.CreateCommanderCommonFlagParams{CommanderID: int64(commanderID), FlagID: int64(flagID)})
}

func ClearCommanderCommonFlag(commanderID uint32, flagID uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.DeleteCommanderCommonFlag(ctx, gen.DeleteCommanderCommonFlagParams{CommanderID: int64(commanderID), FlagID: int64(flagID)})
}
