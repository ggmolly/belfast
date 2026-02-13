package orm

import (
	"context"
	"testing"

	"github.com/ggmolly/belfast/internal/db"
)

func TestListShipsFiltersAndPagination(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Ship{})

	ships := []Ship{
		{TemplateID: 1, Name: "Alpha", EnglishName: "Alpha", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10},
		{TemplateID: 2, Name: "Beta", EnglishName: "Beta", RarityID: 3, Star: 1, Type: 2, Nationality: 2, BuildTime: 10},
		{TemplateID: 3, Name: "Alpha Two", EnglishName: "Alpha Two", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10},
	}
	for i := range ships {
		if err := ships[i].Create(); err != nil {
			t.Fatalf("seed ship: %v", err)
		}
	}
	rarity := uint32(2)
	typeID := uint32(1)
	nat := uint32(1)
	result, err := ListShips(nil, ShipQueryParams{Offset: 0, Limit: 1, RarityID: &rarity, TypeID: &typeID, NationalityID: &nat, Name: "alpha"})
	if err != nil {
		t.Fatalf("list ships: %v", err)
	}
	if result.Total != 2 || len(result.Ships) != 1 {
		t.Fatalf("unexpected ship list result")
	}
}

func TestListItemsResourcesSkins(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Item{})
	clearTable(t, &Resource{})
	clearTable(t, &Skin{})

	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO items (id, name, rarity, shop_id, type, virtual_type) VALUES ($1, $2, $3, $4, $5, $6)`, int64(1), "Item", int64(1), int64(-2), int64(1), int64(0)); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO resources (id, item_id, name) VALUES ($1, $2, $3)`, int64(1), int64(0), "Gold"); err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	if _, err := db.DefaultStore.Pool.Exec(context.Background(), `INSERT INTO skins (id, ship_group, name) VALUES ($1, $2, $3)`, int64(1), int64(10), "Skin"); err != nil {
		t.Fatalf("seed skin: %v", err)
	}

	items, err := ListItems(nil, ItemQueryParams{Offset: 0, Limit: 10})
	if err != nil || items.Total != 1 {
		t.Fatalf("list items: %v", err)
	}
	resources, err := ListResources(nil, ResourceQueryParams{Offset: 0, Limit: 10})
	if err != nil || resources.Total != 1 {
		t.Fatalf("list resources: %v", err)
	}
	skins, err := ListSkins(nil, SkinQueryParams{Offset: 0, Limit: 10})
	if err != nil || skins.Total != 1 {
		t.Fatalf("list skins: %v", err)
	}
	byGroup, err := ListSkinsByShipGroup(nil, 10, SkinQueryParams{Offset: 0, Limit: 10})
	if err != nil || byGroup.Total != 1 {
		t.Fatalf("list skins by group: %v", err)
	}
}
