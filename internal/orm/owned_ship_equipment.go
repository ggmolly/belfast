package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
)

type OwnedShipEquipment struct {
	OwnerID uint32 `gorm:"primaryKey;autoIncrement:false" json:"owner_id"`
	ShipID  uint32 `gorm:"primaryKey;autoIncrement:false" json:"ship_id"`
	Pos     uint32 `gorm:"primaryKey;autoIncrement:false" json:"pos"`
	EquipID uint32 `gorm:"not_null" json:"equip_id"`
	SkinID  uint32 `gorm:"not_null" json:"skin_id"`

	Equipment Equipment `gorm:"foreignKey:EquipID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func ListOwnedShipEquipment(ownerID uint32, shipID uint32) ([]OwnedShipEquipment, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Pool.Query(ctx, `
SELECT owner_id, ship_id, pos, equip_id, skin_id
FROM owned_ship_equipments
WHERE owner_id = $1 AND ship_id = $2
ORDER BY pos ASC
`, int64(ownerID), int64(shipID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entries := make([]OwnedShipEquipment, 0)
	for rows.Next() {
		var row OwnedShipEquipment
		if err := rows.Scan(&row.OwnerID, &row.ShipID, &row.Pos, &row.EquipID, &row.SkinID); err != nil {
			return nil, err
		}
		entries = append(entries, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func GetOwnedShipEquipment(ownerID uint32, shipID uint32, pos uint32) (*OwnedShipEquipment, error) {
	ctx := context.Background()
	entry := OwnedShipEquipment{}
	err := db.DefaultStore.Pool.QueryRow(ctx, `
SELECT owner_id, ship_id, pos, equip_id, skin_id
FROM owned_ship_equipments
WHERE owner_id = $1 AND ship_id = $2 AND pos = $3
`, int64(ownerID), int64(shipID), int64(pos)).Scan(&entry.OwnerID, &entry.ShipID, &entry.Pos, &entry.EquipID, &entry.SkinID)
	err = db.MapNotFound(err)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func UpsertOwnedShipEquipmentTx(ctx context.Context, tx pgx.Tx, entry *OwnedShipEquipment) error {
	_, err := tx.Exec(ctx, `
INSERT INTO owned_ship_equipments (owner_id, ship_id, pos, equip_id, skin_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (owner_id, ship_id, pos)
DO UPDATE SET
  equip_id = EXCLUDED.equip_id,
  skin_id = EXCLUDED.skin_id
`, int64(entry.OwnerID), int64(entry.ShipID), int64(entry.Pos), int64(entry.EquipID), int64(entry.SkinID))
	return err
}
