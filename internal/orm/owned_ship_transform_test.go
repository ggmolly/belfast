package orm

import (
	"context"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
)

func TestConsumeResourceTx(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	clearTransformTable(t, &OwnedResource{})
	clearTransformTable(t, &Commander{})
	commander := Commander{CommanderID: 700, AccountID: 700, Name: "Resource Tester"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO owned_resources (commander_id, resource_id, amount) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(1), int64(10)); err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	ctx := context.Background()
	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return commander.ConsumeResourceTx(ctx, tx, 1, 4)
	}); err != nil {
		t.Fatalf("consume resource: %v", err)
	}
	var resource OwnedResource
	if err := db.DefaultStore.Pool.QueryRow(context.Background(), `SELECT commander_id, resource_id, amount FROM owned_resources WHERE commander_id = $1 AND resource_id = $2`, int64(commander.CommanderID), int64(1)).Scan(&resource.CommanderID, &resource.ResourceID, &resource.Amount); err != nil {
		t.Fatalf("load resource: %v", err)
	}
	if resource.Amount != 6 {
		t.Fatalf("expected resource amount 6, got %d", resource.Amount)
	}
}

func TestToProtoOwnedShipTransformList(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	clearTransformTable(t, &OwnedShipTransform{})
	clearTransformTable(t, &OwnedShip{})
	clearTransformTable(t, &Commander{})
	clearTransformTable(t, &Ship{})
	commander := Commander{CommanderID: 701, AccountID: 701, Name: "Transform Tester"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	ship := Ship{TemplateID: 9001, Name: "Ship", EnglishName: "Ship", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10}
	if err := ship.Create(); err != nil {
		t.Fatalf("create ship: %v", err)
	}
	owned := OwnedShip{OwnerID: commander.CommanderID, ShipID: ship.TemplateID, Level: 1}
	if err := owned.Create(); err != nil {
		t.Fatalf("create owned ship: %v", err)
	}
	transform := OwnedShipTransform{OwnerID: commander.CommanderID, ShipID: owned.ID, TransformID: 12011, Level: 1}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO owned_ship_transforms (owner_id, ship_id, transform_id, level) VALUES ($1, $2, $3, $4)`, int64(transform.OwnerID), int64(transform.ShipID), int64(transform.TransformID), int64(transform.Level)); err != nil {
		t.Fatalf("create transform: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}
	loaded := commander.OwnedShipsMap[owned.ID]
	info := ToProtoOwnedShip(*loaded, nil, nil)
	if len(info.TransformList) != 1 {
		t.Fatalf("expected transform list length 1, got %d", len(info.TransformList))
	}
	if info.TransformList[0].GetId() != 12011 || info.TransformList[0].GetLevel() != 1 {
		t.Fatalf("unexpected transform list entry")
	}
}

func clearTransformTable(t *testing.T, model any) {
	t.Helper()
	clearTable(t, model)
}
