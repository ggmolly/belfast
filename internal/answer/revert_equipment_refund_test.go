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
	clearTable(t, &orm.OwnedEquipment{})
	clearTable(t, &orm.Equipment{})

	if err := orm.GormDB.Create(&orm.Equipment{
		ID:                500,
		Prev:              0,
		Level:             1,
		TransUseGold:      10,
		TransUseItem:      json.RawMessage(`[[200,1]]`),
		ShipTypeForbidden: json.RawMessage(`[]`),
	}).Error; err != nil {
		t.Fatalf("seed root equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.Equipment{
		ID:                501,
		Prev:              500,
		Level:             2,
		TransUseGold:      20,
		TransUseItem:      json.RawMessage(`[[200,2],[201,1]]`),
		ShipTypeForbidden: json.RawMessage(`[]`),
	}).Error; err != nil {
		t.Fatalf("seed mid equipment: %v", err)
	}
	if err := orm.GormDB.Create(&orm.Equipment{
		ID:                502,
		Prev:              501,
		Level:             3,
		TransUseGold:      30,
		TransUseItem:      json.RawMessage(`[[200,3]]`),
		ShipTypeForbidden: json.RawMessage(`[]`),
	}).Error; err != nil {
		t.Fatalf("seed current equipment: %v", err)
	}

	root, items, coins, ok, err := computeRevertEquipmentRefunds(orm.GormDB, 502)
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
