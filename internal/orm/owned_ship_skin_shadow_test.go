package orm

import (
	"sync"
	"testing"
)

var ownedShipSkinShadowTestOnce sync.Once

func initOwnedShipSkinShadowTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	ownedShipSkinShadowTestOnce.Do(func() {
		InitDatabase()
	})
}

func TestUpsertOwnedShipSkinShadowIdempotent(t *testing.T) {
	initOwnedShipSkinShadowTestDB(t)
	commanderID := uint32(410)
	shipID := uint32(420)
	shadowID := uint32(1)

	entry := OwnedShipSkinShadow{CommanderID: commanderID, ShipID: shipID, ShadowID: shadowID, SkinID: 0}
	if err := UpsertOwnedShipSkinShadow(GormDB, &entry); err != nil {
		t.Fatalf("upsert entry: %v", err)
	}
	if err := UpsertOwnedShipSkinShadow(GormDB, &entry); err != nil {
		t.Fatalf("upsert entry second time: %v", err)
	}

	var entries []OwnedShipSkinShadow
	if err := GormDB.Where("commander_id = ? AND ship_id = ? AND shadow_id = ?", commanderID, shipID, shadowID).Find(&entries).Error; err != nil {
		t.Fatalf("load entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestListOwnedShipSkinShadowsBuildsKVData(t *testing.T) {
	initOwnedShipSkinShadowTestDB(t)
	commanderID := uint32(411)
	shipID := uint32(421)

	entries := []OwnedShipSkinShadow{
		{CommanderID: commanderID, ShipID: shipID, ShadowID: 1, SkinID: 0},
		{CommanderID: commanderID, ShipID: shipID, ShadowID: 2, SkinID: 5},
	}
	if err := GormDB.Create(&entries).Error; err != nil {
		t.Fatalf("create entries: %v", err)
	}

	result, err := ListOwnedShipSkinShadows(commanderID, []uint32{shipID})
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	list := result[shipID]
	if len(list) != 2 {
		t.Fatalf("expected 2 kv entries, got %d", len(list))
	}
	if list[0].GetKey() != 1 || list[0].GetValue() != 0 {
		t.Fatalf("unexpected kv0 %d=%d", list[0].GetKey(), list[0].GetValue())
	}
	if list[1].GetKey() != 2 || list[1].GetValue() != 5 {
		t.Fatalf("unexpected kv1 %d=%d", list[1].GetKey(), list[1].GetValue())
	}
}
