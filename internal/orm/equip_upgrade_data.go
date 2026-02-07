package orm

import (
	"encoding/json"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

const equipUpgradeCategory = "ShareCfg/equip_upgrade_data.json"

type EquipUpgradeMaterial struct {
	ItemID uint32
	Count  uint32
}

type EquipUpgradeData struct {
	ID           uint32
	UpgradeFrom  uint32
	TargetID     uint32
	CoinConsume  uint32
	MaterialCost []EquipUpgradeMaterial
}

func GetEquipUpgradeDataTx(db *gorm.DB, upgradeID uint32) (*EquipUpgradeData, error) {
	entry, err := GetConfigEntry(db, equipUpgradeCategory, strconv.FormatUint(uint64(upgradeID), 10))
	if err != nil {
		return nil, err
	}
	var raw struct {
		ID          uint32          `json:"id"`
		UpgradeFrom uint32          `json:"upgrade_from"`
		TargetID    uint32          `json:"target_id"`
		CoinConsume uint32          `json:"coin_consume"`
		Materials   json.RawMessage `json:"material_consume"`
	}
	if err := json.Unmarshal(entry.Data, &raw); err != nil {
		return nil, err
	}
	materials, err := parseEquipUpgradeMaterials(raw.Materials)
	if err != nil {
		return nil, err
	}
	return &EquipUpgradeData{
		ID:           raw.ID,
		UpgradeFrom:  raw.UpgradeFrom,
		TargetID:     raw.TargetID,
		CoinConsume:  raw.CoinConsume,
		MaterialCost: materials,
	}, nil
}

func parseEquipUpgradeMaterials(raw json.RawMessage) ([]EquipUpgradeMaterial, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var pairs [][]uint32
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return nil, err
	}
	materials := make([]EquipUpgradeMaterial, 0, len(pairs))
	for _, pair := range pairs {
		if len(pair) != 2 {
			return nil, errors.New("invalid equip upgrade material_consume")
		}
		materials = append(materials, EquipUpgradeMaterial{ItemID: pair[0], Count: pair[1]})
	}
	return materials, nil
}
