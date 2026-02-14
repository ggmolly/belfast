package orm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
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
		if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO config_entries (category, key, data) VALUES ($1, $2, $3)`, entries[i].Category, entries[i].Key, entries[i].Data); err != nil {
			t.Fatalf("seed config entry: %v", err)
		}
	}
	list, err := ListConfigEntries("alpha")
	if err != nil {
		t.Fatalf("list config entries: %v", err)
	}
	if len(list) != 2 || list[0].Key != "a" || list[1].Key != "b" {
		t.Fatalf("unexpected list order")
	}
	entry, err := GetConfigEntry("beta", "a")
	if err != nil {
		t.Fatalf("get config entry: %v", err)
	}
	if string(entry.Data) != `"three"` {
		t.Fatalf("unexpected entry data")
	}
}
