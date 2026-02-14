package orm

import (
	"context"
	"github.com/ggmolly/belfast/internal/consts"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

func UpdateCommanderRandomShipMode(commanderID uint32, mode uint32) error {
	ctx := context.Background()
	return db.DefaultStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.UpdateCommanderRandomShipMode(ctx, gen.UpdateCommanderRandomShipModeParams{CommanderID: int64(commanderID), RandomShipMode: int64(mode)}); err != nil {
			return err
		}
		if mode == 2 {
			return q.CreateCommanderCommonFlag(ctx, gen.CreateCommanderCommonFlagParams{CommanderID: int64(commanderID), FlagID: int64(consts.RandomFlagShipMode)})
		}
		return q.DeleteCommanderCommonFlag(ctx, gen.DeleteCommanderCommonFlagParams{CommanderID: int64(commanderID), FlagID: int64(consts.RandomFlagShipMode)})
	})
}
