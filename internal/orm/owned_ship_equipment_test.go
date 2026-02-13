package orm

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
)

func TestOwnedShipEquipmentQueries(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedShipEquipment{})
	clearTable(t, &OwnedShip{})
	clearTable(t, &Ship{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 3001, AccountID: 3001, Name: "Ship Equip Owner"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	ship := Ship{TemplateID: 4001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 1}
	if err := ship.Create(); err != nil {
		t.Fatalf("create ship: %v", err)
	}
	owned := OwnedShip{OwnerID: commander.CommanderID, ShipID: ship.TemplateID}
	if err := owned.Create(); err != nil {
		t.Fatalf("create owned ship: %v", err)
	}

	entry := OwnedShipEquipment{OwnerID: commander.CommanderID, ShipID: owned.ID, Pos: 1, EquipID: 9001, SkinID: 0}
	ctx := context.Background()
	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return UpsertOwnedShipEquipmentTx(ctx, tx, &entry)
	}); err != nil {
		t.Fatalf("upsert ship equipment: %v", err)
	}

	loaded, err := GetOwnedShipEquipment(commander.CommanderID, owned.ID, 1)
	if err != nil {
		t.Fatalf("get ship equipment: %v", err)
	}
	if loaded.EquipID != 9001 {
		t.Fatalf("expected equip id 9001, got %d", loaded.EquipID)
	}

	list, err := ListOwnedShipEquipment(commander.CommanderID, owned.ID)
	if err != nil {
		t.Fatalf("list ship equipment: %v", err)
	}
	if len(list) != 1 || list[0].EquipID != 9001 {
		t.Fatalf("expected one equipment entry")
	}
}
