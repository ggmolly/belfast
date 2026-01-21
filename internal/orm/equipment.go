package orm

import "encoding/json"

type Equipment struct {
	ID                uint32          `gorm:"primary_key" json:"id"`
	Base              *uint32         `gorm:"type:int" json:"base"`
	DestroyGold       uint32          `gorm:"not_null" json:"destory_gold"`
	DestroyItem       json.RawMessage `gorm:"type:json" json:"destory_item"`
	EquipLimit        int             `gorm:"not_null" json:"equip_limit"`
	Group             uint32          `gorm:"not_null" json:"group"`
	Important         uint32          `gorm:"not_null" json:"important"`
	Level             uint32          `gorm:"not_null" json:"level"`
	Next              int             `gorm:"not_null" json:"next"`
	Prev              int             `gorm:"not_null" json:"prev"`
	RestoreGold       uint32          `gorm:"not_null" json:"restore_gold"`
	RestoreItem       json.RawMessage `gorm:"type:json" json:"restore_item"`
	ShipTypeForbidden json.RawMessage `gorm:"type:json" json:"ship_type_forbidden"`
	TransUseGold      uint32          `gorm:"not_null" json:"trans_use_gold"`
	TransUseItem      json.RawMessage `gorm:"type:json" json:"trans_use_item"`
	Type              uint32          `gorm:"not_null" json:"type"`
	UpgradeFormulaID  json.RawMessage `gorm:"type:json" json:"upgrade_formula_id"`
}
