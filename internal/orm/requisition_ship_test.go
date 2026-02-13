package orm

import (
	"sync"
	"testing"
)

var requisitionShipTestOnce sync.Once

func initRequisitionShipTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("MODE", "test")
	requisitionShipTestOnce.Do(func() {
		InitDatabase()
	})
}

func TestRequisitionShipQueries(t *testing.T) {
	initRequisitionShipTestDB(t)
	clearTable(t, &RequisitionShip{})
	clearTable(t, &Ship{})

	ships := []Ship{
		{TemplateID: 4101, Name: "Req A", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 1},
		{TemplateID: 4102, Name: "Req B", RarityID: 3, Star: 1, Type: 1, Nationality: 1, BuildTime: 1},
	}
	for i := range ships {
		if err := InsertShip(&ships[i]); err != nil {
			t.Fatalf("create ships: %v", err)
		}
	}
	entries := []RequisitionShip{{ShipID: 4101}, {ShipID: 4102}}
	for i := range entries {
		if err := CreateRequisitionShip(entries[i].ShipID); err != nil {
			t.Fatalf("create requisition entries: %v", err)
		}
	}

	ids, err := ListRequisitionShipIDs()
	if err != nil {
		t.Fatalf("list requisition ids: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 requisition ids, got %d", len(ids))
	}
	lookup := map[uint32]struct{}{4101: {}, 4102: {}}
	for _, id := range ids {
		if _, ok := lookup[id]; !ok {
			t.Fatalf("unexpected requisition id %d", id)
		}
	}

	ship, err := GetRandomRequisitionShipByRarity(2)
	if err != nil {
		t.Fatalf("get random requisition ship: %v", err)
	}
	if ship.RarityID != 2 {
		t.Fatalf("expected rarity 2, got %d", ship.RarityID)
	}
}

func TestListRequisitionShipIDsEmpty(t *testing.T) {
	initRequisitionShipTestDB(t)
	clearTable(t, &RequisitionShip{})
	ids, err := ListRequisitionShipIDs()
	if err != nil {
		t.Fatalf("list requisition ids: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected empty ids")
	}
}
