package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type OwnedShipShadowSkin struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShadowID    uint32 `gorm:"primaryKey;autoIncrement:false"`
	SkinID      uint32 `gorm:"not_null"`
}

func UpsertOwnedShipShadowSkin(_ any, commanderID uint32, shipID uint32, shadowID uint32, skinID uint32) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertOwnedShipShadowSkin(ctx, gen.UpsertOwnedShipShadowSkinParams{
		CommanderID: int64(commanderID),
		ShipID:      int64(shipID),
		ShadowID:    int64(shadowID),
		SkinID:      int64(skinID),
	})
}

func ListOwnedShipShadowSkins(commanderID uint32, shipIDs []uint32) (map[uint32][]OwnedShipShadowSkin, error) {
	ctx := context.Background()
	var entries []gen.OwnedShipShadowSkin
	var err error
	if len(shipIDs) > 0 {
		ids := make([]int64, 0, len(shipIDs))
		for _, shipID := range shipIDs {
			ids = append(ids, int64(shipID))
		}
		entries, err = db.DefaultStore.Queries.ListOwnedShipShadowSkinsByCommanderAndShips(ctx, gen.ListOwnedShipShadowSkinsByCommanderAndShipsParams{CommanderID: int64(commanderID), Column2: ids})
	} else {
		entries, err = db.DefaultStore.Queries.ListOwnedShipShadowSkinsByCommander(ctx, int64(commanderID))
	}
	if err != nil {
		return nil, err
	}
	result := make(map[uint32][]OwnedShipShadowSkin)
	for _, entry := range entries {
		converted := OwnedShipShadowSkin{
			CommanderID: uint32(entry.CommanderID),
			ShipID:      uint32(entry.ShipID),
			ShadowID:    uint32(entry.ShadowID),
			SkinID:      uint32(entry.SkinID),
		}
		result[converted.ShipID] = append(result[converted.ShipID], converted)
	}
	return result, nil
}
