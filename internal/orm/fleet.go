package orm

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/db/gen"
)

var (
	ErrInvalidShipID = errors.New("invalid ship id")
	ErrShipBusy      = errors.New("ship is busy")
)

type Fleet struct {
	ID             uint32    `gorm:"primary_key"` // uniquely identifies the fleet
	GameID         uint32    `gorm:"not_null"`    // uniquely identifies for each commander
	CommanderID    uint32    `gorm:"not_null"`    // owner of the fleet
	Name           string    `gorm:"size:32;not_null;varchar(32)"`
	ShipList       Int64List `gorm:"type:json;not_null"`
	MeowfficerList Int64List `gorm:"type:json;not_null"`
}

// Creates a fleet for the given commander, with the given ships
func CreateFleet(owner *Commander, id uint32, name string, ships []uint32) error {
	var fleet Fleet
	fleet.CommanderID = owner.CommanderID
	fleet.GameID = id
	fleet.Name = name
	busyShipIDs, err := GetBusyEventShipIDs(nil, owner.CommanderID)
	if err != nil {
		return err
	}
	for _, shipID := range ships {
		// check if the commander has this ship
		if _, ok := owner.OwnedShipsMap[shipID]; !ok {
			return ErrInvalidShipID
		}
		if _, ok := busyShipIDs[shipID]; ok {
			return ErrShipBusy
		}
		fleet.ShipList = append(fleet.ShipList, int64(shipID))
	}
	ctx := context.Background()
	shipListJSON, err := json.Marshal(fleet.ShipList)
	if err != nil {
		return err
	}
	meowJSON, err := json.Marshal(fleet.MeowfficerList)
	if err != nil {
		return err
	}
	fleetID, err := db.DefaultStore.Queries.CreateFleet(ctx, gen.CreateFleetParams{
		GameID:         int64(fleet.GameID),
		CommanderID:    int64(fleet.CommanderID),
		Name:           fleet.Name,
		ShipList:       shipListJSON,
		MeowfficerList: meowJSON,
	})
	if err != nil {
		return err
	}
	fleet.ID = uint32(fleetID)
	owner.Fleets = append(owner.Fleets, fleet)
	owner.FleetsMap[fleet.GameID] = &owner.Fleets[len(owner.Fleets)-1]
	return nil
}

func (f *Fleet) RenameFleet(name string) error {
	ctx := context.Background()
	if err := db.DefaultStore.Queries.UpdateFleetName(ctx, gen.UpdateFleetNameParams{ID: int64(f.ID), Name: name}); err != nil {
		return err
	}
	f.Name = name
	return nil
}

// Updates the ship list of the fleet
func (f *Fleet) UpdateShipList(owner *Commander, ships []uint32) error {
	busyShipIDs, err := GetBusyEventShipIDs(nil, owner.CommanderID)
	if err != nil {
		return err
	}
	f.ShipList = make([]int64, len(ships))
	for i, shipID := range ships {
		// check if the commander has this ship
		if _, ok := owner.OwnedShipsMap[shipID]; !ok {
			return ErrInvalidShipID
		}
		if _, ok := busyShipIDs[shipID]; ok {
			return ErrShipBusy
		}
		f.ShipList[i] = int64(shipID)
	}
	ctx := context.Background()
	shipListJSON, err := json.Marshal(f.ShipList)
	if err != nil {
		return err
	}
	if err := db.DefaultStore.Queries.UpdateFleetShipList(ctx, gen.UpdateFleetShipListParams{ID: int64(f.ID), ShipList: shipListJSON}); err != nil {
		return err
	}
	// Update the ship list in the commander's map
	owner.FleetsMap[f.GameID] = f

	// Find the fleet in the commander's list and update it
	for i, fleet := range owner.Fleets {
		if fleet.ID == f.ID {
			owner.Fleets[i] = *f
			break
		}
	}
	return nil
}

func (f *Fleet) AddMeowfficer(shipID uint32) error {
	panic("not implemented")
}

func (f *Fleet) RemoveMeowfficer(shipID uint32) error {
	panic("not implemented")
}

func DeleteFleetByCommanderAndGameID(commanderID uint32, gameID uint32) error {
	ctx := context.Background()
	res, err := db.DefaultStore.Pool.Exec(ctx, `
DELETE FROM fleets
WHERE commander_id = $1
  AND game_id = $2
`, int64(commanderID), int64(gameID))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}
