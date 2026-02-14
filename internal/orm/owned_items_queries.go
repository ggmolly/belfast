package orm

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/ggmolly/belfast/internal/db"
)

func UpsertCommanderMiscItem(commanderID uint32, itemID uint32, data uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO commander_misc_items (commander_id, item_id, data)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET data = EXCLUDED.data
`, int64(commanderID), int64(itemID), int64(data))
	return err
}

func DeleteCommanderItem(commanderID uint32, itemID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_items
WHERE commander_id = $1
  AND item_id = $2
`, int64(commanderID), int64(itemID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteCommanderItemTx(ctx context.Context, tx pgx.Tx, commanderID uint32, itemID uint32) error {
	res, err := tx.Exec(ctx, `
DELETE FROM commander_items
WHERE commander_id = $1
  AND item_id = $2
`, int64(commanderID), int64(itemID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteCommanderMiscItem(commanderID uint32, itemID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM commander_misc_items
WHERE commander_id = $1
  AND item_id = $2
`, int64(commanderID), int64(itemID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteCommanderMiscItemTx(ctx context.Context, tx pgx.Tx, commanderID uint32, itemID uint32) error {
	res, err := tx.Exec(ctx, `
DELETE FROM commander_misc_items
WHERE commander_id = $1
  AND item_id = $2
`, int64(commanderID), int64(itemID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
