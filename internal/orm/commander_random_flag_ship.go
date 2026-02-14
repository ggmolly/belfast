package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func UpdateCommanderRandomFlagShipEnabled(commanderID uint32, enabled bool) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE commanders
SET random_flag_ship_enabled = $2
WHERE commander_id = $1
`, int64(commanderID), enabled)
	return err
}
