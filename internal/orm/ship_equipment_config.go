package orm

import (
	"encoding/json"
	"fmt"
)

const shipDataTemplateCategory = "sharecfgdata/ship_data_template.json"

type ShipEquipConfig struct {
	Equip1 []uint32 `json:"equip_1"`
	Equip2 []uint32 `json:"equip_2"`
	Equip3 []uint32 `json:"equip_3"`
	Equip4 []uint32 `json:"equip_4"`
	Equip5 []uint32 `json:"equip_5"`

	EquipID1 uint32 `json:"equip_id_1"`
	EquipID2 uint32 `json:"equip_id_2"`
	EquipID3 uint32 `json:"equip_id_3"`
}

func GetShipEquipConfig(templateID uint32) (*ShipEquipConfig, error) {
	return GetShipEquipConfigTx(nil, templateID)
}

func GetShipEquipConfigTx(_ any, templateID uint32) (*ShipEquipConfig, error) {
	entry, err := GetConfigEntry(shipDataTemplateCategory, fmt.Sprintf("%d", templateID))
	if err != nil {
		return nil, err
	}
	var config ShipEquipConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *ShipEquipConfig) SlotTypes(pos uint32) []uint32 {
	switch pos {
	case 1:
		return c.Equip1
	case 2:
		return c.Equip2
	case 3:
		return c.Equip3
	case 4:
		return c.Equip4
	case 5:
		return c.Equip5
	default:
		return nil
	}
}

func (c *ShipEquipConfig) SlotCount() uint32 {
	slots := [][]uint32{c.Equip1, c.Equip2, c.Equip3, c.Equip4, c.Equip5}
	var count uint32
	for i, slot := range slots {
		if len(slot) > 0 {
			count = uint32(i + 1)
		}
	}
	if count == 0 {
		return 3
	}
	return count
}

func (c *ShipEquipConfig) DefaultEquipID(pos uint32) uint32 {
	switch pos {
	case 1:
		return c.EquipID1
	case 2:
		return c.EquipID2
	case 3:
		return c.EquipID3
	default:
		return 0
	}
}
