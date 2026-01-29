package orm

import "testing"

func TestListShipsFiltersAndPagination(t *testing.T) {
	initCommanderItemTestDB(t)
	clearTable(t, &Ship{})

	ships := []Ship{
		{TemplateID: 1, Name: "Alpha", EnglishName: "Alpha", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10},
		{TemplateID: 2, Name: "Beta", EnglishName: "Beta", RarityID: 3, Star: 1, Type: 2, Nationality: 2, BuildTime: 10},
		{TemplateID: 3, Name: "Alpha Two", EnglishName: "Alpha Two", RarityID: 2, Star: 1, Type: 1, Nationality: 1, BuildTime: 10},
	}
	for i := range ships {
		if err := GormDB.Create(&ships[i]).Error; err != nil {
			t.Fatalf("seed ship: %v", err)
		}
	}
	rarity := uint32(2)
	typeID := uint32(1)
	nat := uint32(1)
	result, err := ListShips(GormDB, ShipQueryParams{Offset: 0, Limit: 1, RarityID: &rarity, TypeID: &typeID, NationalityID: &nat, Name: "alpha"})
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

	if err := GormDB.Create(&Item{ID: 1, Name: "Item", Rarity: 1, ShopID: -2, Type: 1, VirtualType: 0}).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if err := GormDB.Create(&Resource{ID: 1, Name: "Gold"}).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	if err := GormDB.Create(&Skin{ID: 1, ShipGroup: 10, Name: "Skin"}).Error; err != nil {
		t.Fatalf("seed skin: %v", err)
	}

	items, err := ListItems(GormDB, ItemQueryParams{Offset: 0, Limit: 10})
	if err != nil || items.Total != 1 {
		t.Fatalf("list items: %v", err)
	}
	resources, err := ListResources(GormDB, ResourceQueryParams{Offset: 0, Limit: 10})
	if err != nil || resources.Total != 1 {
		t.Fatalf("list resources: %v", err)
	}
	skins, err := ListSkins(GormDB, SkinQueryParams{Offset: 0, Limit: 10})
	if err != nil || skins.Total != 1 {
		t.Fatalf("list skins: %v", err)
	}
	byGroup, err := ListSkinsByShipGroup(GormDB, 10, SkinQueryParams{Offset: 0, Limit: 10})
	if err != nil || byGroup.Total != 1 {
		t.Fatalf("list skins by group: %v", err)
	}
}
