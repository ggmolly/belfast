package orm

import (
	"encoding/json"
	"testing"
)

func TestLoadSupportRequisitionConfig(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ConfigEntry{})

	data := json.RawMessage(`{"key_value":0,"description":[100,[[2,50],[3,50]],10]}`)
	if err := UpsertConfigEntry("ShareCfg/gameset.json", "supports_config", data); err != nil {
		t.Fatalf("seed config entry: %v", err)
	}
	config, err := LoadSupportRequisitionConfig()
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
	if err := UpsertConfigEntry("ShareCfg/gameset.json", "supports_config", data); err != nil {
		t.Fatalf("seed config entry: %v", err)
	}
	if _, err := LoadSupportRequisitionConfig(); err == nil {
		t.Fatalf("expected error for invalid rarity entry")
	}
}
