package orm

import (
	"encoding/json"
	"testing"
)

func TestGetEquipUpgradeDataTxLoadsConfigEntry(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ConfigEntry{})

	entry := ConfigEntry{
		Category: equipUpgradeCategory,
		Key:      "9001",
		Data:     json.RawMessage(`{"id":9001,"upgrade_from":2001,"target_id":2002,"coin_consume":120,"material_consume":[[3001,2],[3002,4]]}`),
	}
	if err := GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed equip upgrade entry: %v", err)
	}

	data, err := GetEquipUpgradeDataTx(GormDB, 9001)
	if err != nil {
		t.Fatalf("get equip upgrade data: %v", err)
	}
	if data.ID != 9001 || data.UpgradeFrom != 2001 || data.TargetID != 2002 {
		t.Fatalf("unexpected ids: %+v", data)
	}
	if data.CoinConsume != 120 {
		t.Fatalf("expected coin consume 120, got %d", data.CoinConsume)
	}
	if len(data.MaterialCost) != 2 || data.MaterialCost[0].ItemID != 3001 || data.MaterialCost[0].Count != 2 {
		t.Fatalf("unexpected materials: %+v", data.MaterialCost)
	}
}

func TestGetEquipUpgradeDataTxRejectsInvalidMaterials(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ConfigEntry{})

	entry := ConfigEntry{
		Category: equipUpgradeCategory,
		Key:      "9002",
		Data:     json.RawMessage(`{"id":9002,"upgrade_from":2001,"target_id":2002,"coin_consume":0,"material_consume":[[3001]]}`),
	}
	if err := GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed equip upgrade entry: %v", err)
	}

	if _, err := GetEquipUpgradeDataTx(GormDB, 9002); err == nil {
		t.Fatalf("expected error")
	}
}
