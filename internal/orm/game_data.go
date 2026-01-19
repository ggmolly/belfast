package orm

import (
	"strings"

	"gorm.io/gorm"
)

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

func ListShips(db *gorm.DB, params ShipQueryParams) (ShipListResult, error) {
	query := db.Model(&Ship{})
	query = applyShipFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ShipListResult{}, err
	}

	var ships []Ship
	if err := query.
		Order("template_id asc").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&ships).Error; err != nil {
		return ShipListResult{}, err
	}

	return ShipListResult{Ships: ships, Total: total}, nil
}

func ListItems(db *gorm.DB, params ItemQueryParams) (ItemListResult, error) {
	query := db.Model(&Item{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ItemListResult{}, err
	}

	var items []Item
	if err := query.
		Order("id asc").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&items).Error; err != nil {
		return ItemListResult{}, err
	}

	return ItemListResult{Items: items, Total: total}, nil
}

func ListResources(db *gorm.DB, params ResourceQueryParams) (ResourceListResult, error) {
	query := db.Model(&Resource{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return ResourceListResult{}, err
	}

	var resources []Resource
	if err := query.
		Order("id asc").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&resources).Error; err != nil {
		return ResourceListResult{}, err
	}

	return ResourceListResult{Resources: resources, Total: total}, nil
}

func ListSkins(db *gorm.DB, params SkinQueryParams) (SkinListResult, error) {
	query := db.Model(&Skin{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return SkinListResult{}, err
	}

	var skins []Skin
	if err := query.
		Order("id asc").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&skins).Error; err != nil {
		return SkinListResult{}, err
	}

	return SkinListResult{Skins: skins, Total: total}, nil
}

func ListSkinsByShipGroup(db *gorm.DB, shipGroup uint32, params SkinQueryParams) (SkinListResult, error) {
	query := db.Model(&Skin{}).Where("ship_group = ?", shipGroup)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return SkinListResult{}, err
	}

	var skins []Skin
	if err := query.
		Order("id asc").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&skins).Error; err != nil {
		return SkinListResult{}, err
	}

	return SkinListResult{Skins: skins, Total: total}, nil
}

func applyShipFilters(query *gorm.DB, params ShipQueryParams) *gorm.DB {
	if params.RarityID != nil {
		query = query.Where("rarity_id = ?", *params.RarityID)
	}
	if params.TypeID != nil {
		query = query.Where("type = ?", *params.TypeID)
	}
	if params.NationalityID != nil {
		query = query.Where("nationality = ?", *params.NationalityID)
	}
	if strings.TrimSpace(params.Name) != "" {
		search := strings.ToLower(strings.TrimSpace(params.Name))
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}
	return query
}
