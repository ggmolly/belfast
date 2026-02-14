package orm

import "strings"

type ShipQueryParams struct {
	Offset        int
	Limit         int
	RarityID      *uint32
	TypeID        *uint32
	NationalityID *uint32
	Name          string
}

type ShipListResult struct {
	Ships []Ship
	Total int64
}

type ItemQueryParams struct {
	Offset int
	Limit  int
}

type ItemListResult struct {
	Items []Item
	Total int64
}

type ResourceQueryParams struct {
	Offset int
	Limit  int
}

type ResourceListResult struct {
	Resources []Resource
	Total     int64
}

type SkinQueryParams struct {
	Offset int
	Limit  int
}

type SkinListResult struct {
	Skins []Skin
	Total int64
}

func ListShips(_ any, params ShipQueryParams) (ShipListResult, error) {
	params.Name = strings.TrimSpace(params.Name)
	ships, total, err := ListShipsPage(params)
	if err != nil {
		return ShipListResult{}, err
	}
	return ShipListResult{Ships: ships, Total: total}, nil
}

func ListItems(_ any, params ItemQueryParams) (ItemListResult, error) {
	items, total, err := ListItemsPage(params.Offset, params.Limit)
	if err != nil {
		return ItemListResult{}, err
	}
	return ItemListResult{Items: items, Total: total}, nil
}

func ListResources(_ any, params ResourceQueryParams) (ResourceListResult, error) {
	resources, total, err := ListResourcesPage(params.Offset, params.Limit)
	if err != nil {
		return ResourceListResult{}, err
	}
	return ResourceListResult{Resources: resources, Total: total}, nil
}

func ListSkins(_ any, params SkinQueryParams) (SkinListResult, error) {
	skins, total, err := ListSkinsPage(params.Offset, params.Limit)
	if err != nil {
		return SkinListResult{}, err
	}
	return SkinListResult{Skins: skins, Total: total}, nil
}

func ListSkinsByShipGroup(_ any, shipGroup uint32, params SkinQueryParams) (SkinListResult, error) {
	skins, total, err := ListSkinsByShipGroupPage(shipGroup, params.Offset, params.Limit)
	if err != nil {
		return SkinListResult{}, err
	}
	return SkinListResult{Skins: skins, Total: total}, nil
}
