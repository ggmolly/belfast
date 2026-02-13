package orm

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/jackc/pgx/v5"
)

func TestOwnedEquipmentSetAndRemove(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedEquipment{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 2001, AccountID: 2001, Name: "Equip Owner"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	ctx := context.Background()
	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return commander.SetOwnedEquipmentTx(ctx, tx, 3001, 2)
	}); err != nil {
		t.Fatalf("set owned equipment: %v", err)
	}

	entry := commander.GetOwnedEquipment(3001)
	if entry == nil || entry.Count != 2 {
		t.Fatalf("expected equipment count 2, got %v", entry)
	}

	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return commander.RemoveOwnedEquipmentTx(ctx, tx, 3001, 1)
	}); err != nil {
		t.Fatalf("remove owned equipment: %v", err)
	}
	entry = commander.GetOwnedEquipment(3001)
	if entry == nil || entry.Count != 1 {
		t.Fatalf("expected equipment count 1, got %v", entry)
	}

	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return commander.SetOwnedEquipmentTx(ctx, tx, 3001, 0)
	}); err != nil {
		t.Fatalf("delete owned equipment: %v", err)
	}
	if commander.GetOwnedEquipment(3001) != nil {
		t.Fatalf("expected equipment to be deleted")
	}
}

func TestOwnedEquipmentMapAfterSliceMutation(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedEquipment{})
	clearTable(t, &Commander{})

	commander := Commander{CommanderID: 2002, AccountID: 2002, Name: "Equip Owner"}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO commanders (commander_id, account_id, name) VALUES ($1, $2, $3)`, int64(commander.CommanderID), int64(commander.AccountID), commander.Name); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	if err := commander.Load(); err != nil {
		t.Fatalf("load commander: %v", err)
	}

	ctx := context.Background()
	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		if err := commander.SetOwnedEquipmentTx(ctx, tx, 4001, 1); err != nil {
			return err
		}
		return commander.SetOwnedEquipmentTx(ctx, tx, 4002, 2)
	}); err != nil {
		t.Fatalf("seed equipment: %v", err)
	}

	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return commander.SetOwnedEquipmentTx(ctx, tx, 4001, 0)
	}); err != nil {
		t.Fatalf("delete equipment: %v", err)
	}

	if err := WithPGXTx(ctx, func(tx pgx.Tx) error {
		return commander.SetOwnedEquipmentTx(ctx, tx, 4002, 5)
	}); err != nil {
		t.Fatalf("update equipment: %v", err)
	}

	entry := commander.GetOwnedEquipment(4002)
	if entry == nil || entry.Count != 5 {
		t.Fatalf("expected map count 5, got %v", entry)
	}
	if count := ownedEquipmentCount(commander.OwnedEquipments, 4002); count != 5 {
		t.Fatalf("expected slice count 5, got %d", count)
	}
}

func ownedEquipmentCount(entries []OwnedEquipment, equipmentID uint32) uint32 {
	for _, entry := range entries {
		if entry.EquipmentID == equipmentID {
			return entry.Count
		}
	}
	return 0
}
