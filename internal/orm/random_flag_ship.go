package orm

import (
	"context"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

type RandomFlagShip struct {
	CommanderID uint32 `gorm:"primaryKey;autoIncrement:false"`
	ShipID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	PhantomID   uint32 `gorm:"primaryKey;autoIncrement:false"`
	Enabled     bool   `gorm:"default:true;not_null"`
}

type RandomFlagShipUpdate struct {
	ShipID    uint32
	PhantomID uint32
	Flag      uint32
}

func ApplyRandomFlagShipUpdates(_ any, commanderID uint32, updates []RandomFlagShipUpdate) error {
	ctx := context.Background()
	for _, update := range updates {
		if update.Flag == 0 {
			if err := db.DefaultStore.Queries.DeleteRandomFlagShip(ctx, gen.DeleteRandomFlagShipParams{CommanderID: int64(commanderID), ShipID: int64(update.ShipID), PhantomID: int64(update.PhantomID)}); err != nil {
				return err
			}
			continue
		}
		if err := db.DefaultStore.Queries.UpsertRandomFlagShip(ctx, gen.UpsertRandomFlagShipParams{CommanderID: int64(commanderID), ShipID: int64(update.ShipID), PhantomID: int64(update.PhantomID), Enabled: true}); err != nil {
			return err
		}
	}
	return nil
}

func ListRandomFlagShipPhantoms(commanderID uint32, shipIDs []uint32) (map[uint32][]uint32, error) {
	flags := make(map[uint32][]uint32)
	ctx := context.Background()
	var entries []gen.RandomFlagShip
	var err error
	if len(shipIDs) > 0 {
		ids := make([]int64, 0, len(shipIDs))
		for _, shipID := range shipIDs {
			ids = append(ids, int64(shipID))
		}
		entries, err = db.DefaultStore.Queries.ListEnabledRandomFlagShipsByCommanderAndShips(ctx, gen.ListEnabledRandomFlagShipsByCommanderAndShipsParams{CommanderID: int64(commanderID), Column2: ids})
	} else {
		entries, err = db.DefaultStore.Queries.ListEnabledRandomFlagShipsByCommander(ctx, int64(commanderID))
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		shipID := uint32(entry.ShipID)
		flags[shipID] = append(flags[shipID], uint32(entry.PhantomID))
	}
	return flags, nil
}

func ListRandomFlagShips(commanderID uint32, shipID *uint32) ([]RandomFlagShip, error) {
	ctx := context.Background()
	query := `
SELECT commander_id, ship_id, phantom_id, enabled
FROM random_flag_ships
WHERE commander_id = $1
`
	args := []any{int64(commanderID)}
	if shipID != nil {
		query += ` AND ship_id = $2`
		args = append(args, int64(*shipID))
	}
	query += ` ORDER BY ship_id ASC, phantom_id ASC`

	rows, err := db.DefaultStore.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]RandomFlagShip, 0)
	for rows.Next() {
		var entry RandomFlagShip
		if err := rows.Scan(&entry.CommanderID, &entry.ShipID, &entry.PhantomID, &entry.Enabled); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func UpsertRandomFlagShipEntry(entry RandomFlagShip) error {
	ctx := context.Background()
	return db.DefaultStore.Queries.UpsertRandomFlagShip(ctx, gen.UpsertRandomFlagShipParams{
		CommanderID: int64(entry.CommanderID),
		ShipID:      int64(entry.ShipID),
		PhantomID:   int64(entry.PhantomID),
		Enabled:     entry.Enabled,
	})
}

func DeleteRandomFlagShipEntry(commanderID uint32, shipID uint32, phantomID uint32) error {
	ctx := context.Background()
	_, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM random_flag_ships
WHERE commander_id = $1
  AND ship_id = $2
  AND phantom_id = $3
`, int64(commanderID), int64(shipID), int64(phantomID))
	return err
}
