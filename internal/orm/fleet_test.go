package orm

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestFleetCreateAndUpdate(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Fleet{})
	clearTable(t, &Commander{})
	clearTable(t, &OwnedShip{})

	commander := Commander{CommanderID: 300, AccountID: 300, Name: "Fleet"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	commander.OwnedShipsMap = make(map[uint32]*OwnedShip)
	commander.FleetsMap = make(map[uint32]*Fleet)

	ship := Ship{TemplateID: 100, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 1}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed ship: %v", err)
	}

	owned := OwnedShip{OwnerID: commander.CommanderID, ShipID: ship.TemplateID}
	if err := owned.Create(); err != nil {
		t.Fatalf("seed owned ship: %v", err)
	}
	commander.OwnedShipsMap[owned.ID] = &owned

	if err := CreateFleet(&commander, 1, "Main", []uint32{owned.ID}); err != nil {
		t.Fatalf("create fleet: %v", err)
	}
	if len(commander.Fleets) != 1 {
		t.Fatalf("expected fleet created")
	}
	fleet := &commander.Fleets[0]
	if err := fleet.RenameFleet("Updated"); err != nil {
		t.Fatalf("rename fleet: %v", err)
	}
	if err := fleet.UpdateShipList(&commander, []uint32{owned.ID}); err != nil {
		t.Fatalf("update ship list: %v", err)
	}
	if err := fleet.UpdateShipList(&commander, []uint32{999}); err == nil {
		t.Fatalf("expected invalid ship id error")
	}

	if err := CreateFleet(&commander, 2, "Bad", []uint32{999}); err == nil {
		t.Fatalf("expected invalid ship id error")
	}
}

func TestFleetMeowfficerPanics(t *testing.T) {
	fleet := Fleet{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from add meowfficer")
		}
	}()
	_ = fleet.AddMeowfficer(1)
}

func TestFleetRemoveMeowfficerPanics(t *testing.T) {
	fleet := Fleet{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic from remove meowfficer")
		}
	}()
	_ = fleet.RemoveMeowfficer(1)
}
