package orm

import (
	"context"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func ListSkinNamesByIDs(skinIDs []uint32) (map[uint32]string, error) {
	result := make(map[uint32]string, len(skinIDs))
	if len(skinIDs) == 0 {
		return result, nil
	}
	ids := make([]int64, 0, len(skinIDs))
	for _, id := range skinIDs {
		ids = append(ids, int64(id))
	}
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT id, name
FROM skins
WHERE id = ANY($1::bigint[])
`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uint32
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		result[id] = name
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func GetSkinNameByID(skinID uint32) (string, error) {
	skin, err := GetSkinByID(skinID)
	if err != nil {
		return "", err
	}
	return skin.Name, nil
}

func UpdateOwnedSkinExpiry(commanderID uint32, skinID uint32, expiresAt *time.Time) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
UPDATE owned_skins
SET expires_at = $3
WHERE commander_id = $1
  AND skin_id = $2
`, int64(commanderID), int64(skinID), expiresAt)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func DeleteOwnedSkin(commanderID uint32, skinID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM owned_skins
WHERE commander_id = $1
  AND skin_id = $2
`, int64(commanderID), int64(skinID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
