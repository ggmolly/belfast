package orm

import (
	"encoding/json"
	"testing"
)

func TestLoadSupportRequisitionConfig(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ConfigEntry{})

	data := json.RawMessage(`{"key_value":0,"description":[100,[[2,50],[3,50]],10]}`)
	entry := ConfigEntry{Category: "ShareCfg/gameset.json", Key: "supports_config", Data: data}
	if err := GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry: %v", err)
	}
	config, err := LoadSupportRequisitionConfig(GormDB)
	if err != nil {
		t.Fatalf("load support requisition config: %v", err)
	}
	if config.Cost != 100 || config.MonthlyCap != 10 || len(config.RarityWeights) != 2 {
		t.Fatalf("unexpected config: %+v", config)
	}
}

func TestLoadSupportRequisitionConfigErrors(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ConfigEntry{})

	data := json.RawMessage(`{"key_value":0,"description":[100,[[2]],10]}`)
	entry := ConfigEntry{Category: "ShareCfg/gameset.json", Key: "supports_config", Data: data}
	if err := GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed config entry: %v", err)
	}
	if _, err := LoadSupportRequisitionConfig(GormDB); err == nil {
		t.Fatalf("expected error for invalid rarity entry")
	}
}
