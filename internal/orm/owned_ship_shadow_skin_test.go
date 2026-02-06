package orm

import "testing"

func TestOwnedShipShadowSkinUpsertAndList(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &OwnedShipShadowSkin{})

	tx := GormDB.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}
	if err := UpsertOwnedShipShadowSkin(tx, 1, 10, 2, 100); err != nil {
		tx.Rollback()
		t.Fatalf("upsert: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	tx = GormDB.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}
	if err := UpsertOwnedShipShadowSkin(tx, 1, 10, 2, 200); err != nil {
		tx.Rollback()
		t.Fatalf("upsert update: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	result, err := ListOwnedShipShadowSkins(1, []uint32{10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	entries := result[10]
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ShadowID != 2 || entries[0].SkinID != 200 {
		t.Fatalf("unexpected entry: %+v", entries[0])
	}
}
