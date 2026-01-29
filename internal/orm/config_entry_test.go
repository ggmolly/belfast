package orm

import (
	"encoding/json"
	"testing"
)

func TestConfigEntryListAndGet(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &ConfigEntry{})

	entries := []ConfigEntry{
		{Category: "alpha", Key: "a", Data: json.RawMessage(`"one"`)},
		{Category: "alpha", Key: "b", Data: json.RawMessage(`"two"`)},
		{Category: "beta", Key: "a", Data: json.RawMessage(`"three"`)},
	}
	for i := range entries {
		if err := GormDB.Create(&entries[i]).Error; err != nil {
			t.Fatalf("seed config entry: %v", err)
		}
	}
	list, err := ListConfigEntries(GormDB, "alpha")
	if err != nil {
		t.Fatalf("list config entries: %v", err)
	}
	if len(list) != 2 || list[0].Key != "a" || list[1].Key != "b" {
		t.Fatalf("unexpected list order")
	}
	entry, err := GetConfigEntry(GormDB, "beta", "a")
	if err != nil {
		t.Fatalf("get config entry: %v", err)
	}
	if string(entry.Data) != `"three"` {
		t.Fatalf("unexpected entry data")
	}
}
