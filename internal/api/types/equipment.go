package types

type EquipmentPayload struct {
	ID                uint32  `json:"id"`
	Base              *uint32 `json:"base"`
	DestroyGold       uint32  `json:"destory_gold"`
	DestroyItem       RawJSON `json:"destory_item"`
	EquipLimit        int     `json:"equip_limit"`
	Group             uint32  `json:"group"`
	Important         uint32  `json:"important"`
	Level             uint32  `json:"level"`
	Next              int     `json:"next"`
	Prev              int     `json:"prev"`
	RestoreGold       uint32  `json:"restore_gold"`
	RestoreItem       RawJSON `json:"restore_item"`
	ShipTypeForbidden RawJSON `json:"ship_type_forbidden"`
	TransUseGold      uint32  `json:"trans_use_gold"`
	TransUseItem      RawJSON `json:"trans_use_item"`
	Type              uint32  `json:"type"`
	UpgradeFormulaID  RawJSON `json:"upgrade_formula_id"`
}

type EquipmentListResponse struct {
	Equipment []EquipmentPayload `json:"equipment"`
	Meta      PaginationMeta     `json:"meta"`
}
