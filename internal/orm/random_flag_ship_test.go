package orm

import (
	"reflect"
	"sync"
	"testing"
)

var randomFlagShipTestOnce sync.Once

func initRandomFlagShipTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	randomFlagShipTestOnce.Do(func() {
		InitDatabase()
	})
}

func TestApplyRandomFlagShipUpdates(t *testing.T) {
	initRandomFlagShipTestDB(t)
	commanderID := uint32(100)
	updates := []RandomFlagShipUpdate{
		{ShipID: 1000, PhantomID: 1, Flag: 1},
		{ShipID: 1000, PhantomID: 2, Flag: 1},
	}
	tx := GormDB.Begin()
	if err := ApplyRandomFlagShipUpdates(tx, commanderID, updates); err != nil {
		t.Fatalf("apply updates: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit updates: %v", err)
	}
	var entries []RandomFlagShip
	if err := GormDB.Where("commander_id = ?", commanderID).Find(&entries).Error; err != nil {
		t.Fatalf("load entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	deleteUpdate := []RandomFlagShipUpdate{{ShipID: 1000, PhantomID: 1, Flag: 0}}
	tx = GormDB.Begin()
	if err := ApplyRandomFlagShipUpdates(tx, commanderID, deleteUpdate); err != nil {
		t.Fatalf("apply delete: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit delete: %v", err)
	}
	entries = entries[:0]
	if err := GormDB.Where("commander_id = ?", commanderID).Find(&entries).Error; err != nil {
		t.Fatalf("load after delete: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after delete, got %d", len(entries))
	}
}

func TestListRandomFlagShipPhantoms(t *testing.T) {
	initRandomFlagShipTestDB(t)
	commanderID := uint32(200)
	entries := []RandomFlagShip{
		{CommanderID: commanderID, ShipID: 2000, PhantomID: 0, Enabled: true},
		{CommanderID: commanderID, ShipID: 2000, PhantomID: 2, Enabled: true},
		{CommanderID: commanderID, ShipID: 2001, PhantomID: 1, Enabled: true},
	}
	if err := GormDB.Create(&entries).Error; err != nil {
		t.Fatalf("create entries: %v", err)
	}
	flags, err := ListRandomFlagShipPhantoms(commanderID, []uint32{2000, 2001})
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	if len(flags[2000]) != 2 {
		t.Fatalf("expected 2 flags for ship 2000, got %d", len(flags[2000]))
	}
	if len(flags[2001]) != 1 {
		t.Fatalf("expected 1 flag for ship 2001, got %d", len(flags[2001]))
	}
}

func TestListRandomFlagShipPhantomsAll(t *testing.T) {
	initRandomFlagShipTestDB(t)
	commanderID := uint32(201)
	entries := []RandomFlagShip{{CommanderID: commanderID, ShipID: 2100, PhantomID: 1, Enabled: true}}
	if err := GormDB.Create(&entries).Error; err != nil {
		t.Fatalf("create entries: %v", err)
	}
	flags, err := ListRandomFlagShipPhantoms(commanderID, nil)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	if len(flags[2100]) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags[2100]))
	}
}

func TestToProtoOwnedShipListRandomFlags(t *testing.T) {
	ship := OwnedShip{ID: 3000, ShipID: 300}
	flags := map[uint32][]uint32{
		3000: {0, 3},
	}
	result := ToProtoOwnedShipList([]OwnedShip{ship}, flags)
	if len(result) != 1 {
		t.Fatalf("expected 1 ship, got %d", len(result))
	}
	if got := result[0].GetCharRandomFlag(); !reflect.DeepEqual(got, []uint32{0, 3}) {
		t.Fatalf("expected char_random_flag [0 3], got %v", got)
	}
}
