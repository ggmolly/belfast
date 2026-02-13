package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

func CommanderExists(commanderID uint32) error {
	ctx := context.Background()
	var id int64
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT commander_id
FROM commanders
WHERE commander_id = $1
`, int64(commanderID)).Scan(&id)
	err = db.MapNotFound(err)
	if err != nil {
		return err
	}
	return nil
}
