package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
)

type OwnedShipStrength struct {
	OwnerID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID     uint32 `gorm:"primaryKey;autoIncrement:false"`
	StrengthID uint32 `gorm:"primaryKey;autoIncrement:false"`
	Exp        uint32 `gorm:"not_null"`
}

func ListOwnedShipStrengths(ownerID uint32, shipID uint32) ([]OwnedShipStrength, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT owner_id, ship_id, strength_id, exp
FROM owned_ship_strengths
WHERE owner_id = $1 AND ship_id = $2
ORDER BY strength_id ASC
`, int64(ownerID), int64(shipID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]OwnedShipStrength, 0)
	for rows.Next() {
		var row OwnedShipStrength
		if err := rows.Scan(&row.OwnerID, &row.ShipID, &row.StrengthID, &row.Exp); err != nil {
			return nil, err
		}
		entries = append(entries, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func UpsertOwnedShipStrengthTx(ctx context.Context, tx pgx.Tx, entry *OwnedShipStrength) error {
	_, err := tx.Exec(ctx, `
INSERT INTO owned_ship_strengths (owner_id, ship_id, strength_id, exp)
VALUES ($1, $2, $3, $4)
ON CONFLICT (owner_id, ship_id, strength_id)
DO UPDATE SET exp = EXCLUDED.exp
`, int64(entry.OwnerID), int64(entry.ShipID), int64(entry.StrengthID), int64(entry.Exp))
	return err
}
