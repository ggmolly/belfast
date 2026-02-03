package orm

import (
	"encoding/json"
	"os"
	"testing"

	"gorm.io/gorm"
)

func TestGetShipBreakoutConfig(t *testing.T) {
	os.Setenv("MODE", "test")
	InitDatabase()
	if err := GormDB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&ConfigEntry{}).Error; err != nil {
		t.Fatalf("clear config entries: %v", err)
	}
	entry := ConfigEntry{
		Category: shipBreakoutCategory,
		Key:      "101021",
		Data: json.RawMessage(`{
"id":101021,
"breakout_id":101022,
"pre_id":0,
"level":10,
"use_gold":300,
"use_item":[[17001,2]],
"use_char":10102,
"use_char_num":1,
"weapon_ids":[70011],
"breakout_view":"Unlock All Out Assault"
}`),
	}
	if err := GormDB.Create(&entry).Error; err != nil {
		t.Fatalf("seed breakout config: %v", err)
	}
	config, err := GetShipBreakoutConfig(101021)
	if err != nil {
		t.Fatalf("GetShipBreakoutConfig failed: %v", err)
	}
	if config.BreakoutID != 101022 {
		t.Fatalf("expected breakout_id 101022, got %d", config.BreakoutID)
	}
	if config.UseChar != 10102 || config.UseCharNum != 1 {
		t.Fatalf("unexpected use_char values")
	}
	if config.UseGold != 300 {
		t.Fatalf("expected use_gold 300, got %d", config.UseGold)
	}
	if len(config.UseItem) != 1 || len(config.UseItem[0]) != 2 || config.UseItem[0][0] != 17001 || config.UseItem[0][1] != 2 {
		t.Fatalf("unexpected use_item data")
	}
}
