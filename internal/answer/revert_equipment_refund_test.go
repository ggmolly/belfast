package answer

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
)

func TestComputeRevertEquipmentRefundsUsesPrevTransUse(t *testing.T) {
	os.Setenv("MODE", "test")
	orm.InitDatabase()
	clearTable(t, &orm.ConfigEntry{})

	if err := orm.UpsertConfigEntry(
		"sharecfgdata/equip_data_statistics.json",
		"500",
		json.RawMessage(`{"id":500,"prev":0,"level":1,"trans_use_gold":10,"trans_use_item":[[200,1]]}`),
	); err != nil {
		t.Fatalf("seed root equipment: %v", err)
	}
	if err := orm.UpsertConfigEntry(
		"sharecfgdata/equip_data_statistics.json",
		"501",
		json.RawMessage(`{"id":501,"prev":500,"level":2,"trans_use_gold":20,"trans_use_item":[[200,2],[201,1]]}`),
	); err != nil {
		t.Fatalf("seed mid equipment: %v", err)
	}
	if err := orm.UpsertConfigEntry(
		"sharecfgdata/equip_data_statistics.json",
		"502",
		json.RawMessage(`{"id":502,"prev":501,"level":3,"trans_use_gold":30,"trans_use_item":[[200,3]]}`),
	); err != nil {
		t.Fatalf("seed current equipment: %v", err)
	}

	root, items, coins, ok, err := computeRevertEquipmentRefunds(502)
	if err != nil {
		t.Fatalf("compute refunds: %v", err)
	}
	if !ok {
		t.Fatalf("expected equip 502 to be revertable")
	}
	if root != 500 {
		t.Fatalf("expected root 500, got %d", root)
	}
	if coins != 30 {
		t.Fatalf("expected refund coins 30, got %d", coins)
	}
	if items[200] != 3 || items[201] != 1 {
		t.Fatalf("expected refund items map[200]=3 map[201]=1, got %#v", items)
	}
}
