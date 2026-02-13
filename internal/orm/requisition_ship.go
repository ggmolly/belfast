package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
)

type RequisitionShip struct {
	ShipID uint32 `gorm:"primary_key" json:"ship_id"`
}

func ListRequisitionShipIDs() ([]uint32, error) {
	ctx := context.Background()
	rows, err := db.DefaultStore.Queries.ListRequisitionShipIDs(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uint32, len(rows))
	for i, id := range rows {
		ids[i] = uint32(id)
	}
	return ids, nil
}

func CreateRequisitionShip(shipID uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
INSERT INTO requisition_ships (ship_id)
VALUES ($1)
`, int64(shipID))
	return err
}

func DeleteRequisitionShip(shipID uint32) error {
	ctx := context.Background()
	tag, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM requisition_ships
WHERE ship_id = $1
`, int64(shipID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

func GetRandomRequisitionShipByRarity(rarity uint32) (Ship, error) {
	ctx := context.Background()
	row, err := db.DefaultStore.Queries.GetRandomRequisitionShipByRarity(ctx, int64(rarity))
	err = db.MapNotFound(err)
	if err != nil {
		return Ship{}, err
	}
	var poolID *uint32
	if row.PoolID.Valid {
		v := uint32(row.PoolID.Int64)
		poolID = &v
	}
	ship := Ship{
		TemplateID:  uint32(row.TemplateID),
		Name:        row.Name,
		EnglishName: row.EnglishName,
		RarityID:    uint32(row.RarityID),
		Star:        uint32(row.Star),
		Type:        uint32(row.Type),
		Nationality: uint32(row.Nationality),
		BuildTime:   uint32(row.BuildTime),
		PoolID:      poolID,
	}
	return ship, nil
}
