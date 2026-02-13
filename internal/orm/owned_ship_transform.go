package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
)

type OwnedShipTransform struct {
	OwnerID     uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	TransformID uint32 `gorm:"primaryKey;autoIncrement:false"`
	Level       uint32 `gorm:"not_null"`
}

func ListOwnedShipTransforms(ownerID uint32, shipID uint32) ([]OwnedShipTransform, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT owner_id, ship_id, transform_id, level
FROM owned_ship_transforms
WHERE owner_id = $1 AND ship_id = $2
ORDER BY transform_id ASC
`, int64(ownerID), int64(shipID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := make([]OwnedShipTransform, 0)
	for rows.Next() {
		var row OwnedShipTransform
		if err := rows.Scan(&row.OwnerID, &row.ShipID, &row.TransformID, &row.Level); err != nil {
			return nil, err
		}
		entries = append(entries, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func UpsertOwnedShipTransformTx(ctx context.Context, tx pgx.Tx, entry *OwnedShipTransform) error {
	_, err := tx.Exec(ctx, `
INSERT INTO owned_ship_transforms (owner_id, ship_id, transform_id, level)
VALUES ($1, $2, $3, $4)
ON CONFLICT (owner_id, ship_id, transform_id)
DO UPDATE SET level = EXCLUDED.level
`, int64(entry.OwnerID), int64(entry.ShipID), int64(entry.TransformID), int64(entry.Level))
	return err
}

func DeleteOwnedShipTransformsTx(ctx context.Context, tx pgx.Tx, ownerID uint32, shipID uint32, transformIDs []uint32) error {
	if len(transformIDs) == 0 {
		return nil
	}
	ids := make([]int64, len(transformIDs))
	for i, id := range transformIDs {
		ids[i] = int64(id)
	}
	_, err := tx.Exec(ctx, `
DELETE FROM owned_ship_transforms
WHERE owner_id = $1 AND ship_id = $2 AND transform_id = ANY($3)
`, int64(ownerID), int64(shipID), ids)
	return err
}
