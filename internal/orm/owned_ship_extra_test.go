package orm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ggmolly/belfast/internal/db"
)

func TestOwnedShipCRUDAndFlags(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 50, AccountID: 50, Name: "Owner"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("seed commander: %v", err)
	}
	ship := Ship{TemplateID: 3001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("seed ship: %v", err)
	}
	owned := OwnedShip{OwnerID: commander.CommanderID, ShipID: ship.TemplateID}
	if err := owned.Create(); err != nil {
		t.Fatalf("create owned ship: %v", err)
	}
	if err := owned.SetFavorite(1); err != nil {
		t.Fatalf("set favorite: %v", err)
	}
	if !owned.CommonFlag {
		t.Fatalf("expected common flag true")
	}
	if err := owned.ProposeShip(); err != nil {
		t.Fatalf("propose ship: %v", err)
	}
	if !owned.Propose {
		t.Fatalf("expected propose true")
	}
	old := time.Now().Add(-40 * 24 * time.Hour)
	owned.ChangeNameTimestamp = old
	if err := owned.RenameShip("Renamed"); err != nil {
		t.Fatalf("rename ship: %v", err)
	}
	if owned.CustomName != "Renamed" {
		t.Fatalf("expected renamed ship")
	}
	if err := owned.Update(); err != nil {
		t.Fatalf("update owned ship: %v", err)
	}
	if err := owned.Delete(); err != nil {
		t.Fatalf("delete owned ship: %v", err)
	}
}

func TestOwnedShipRenameErrors(t *testing.T) {
	owned := OwnedShip{Propose: false}
	if err := owned.RenameShip("Name"); !errors.Is(err, ErrNotProposed) {
		t.Fatalf("expected ErrNotProposed, got %v", err)
	}
	owned.Propose = true
	owned.ChangeNameTimestamp = time.Now()
	if err := owned.RenameShip("Name"); !errors.Is(err, ErrRenameInCooldown) {
		t.Fatalf("expected ErrRenameInCooldown, got %v", err)
	}
}
