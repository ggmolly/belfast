package orm

import (
	"encoding/json"
	"fmt"
)

const shipBreakoutCategory = "sharecfgdata/ship_data_breakout.json"

type ShipBreakoutConfig struct {
	ID           uint32     `json:"id"`
	BreakoutID   uint32     `json:"breakout_id"`
	PreID        uint32     `json:"pre_id"`
	Level        uint32     `json:"level"`
	UseGold      uint32     `json:"use_gold"`
	UseItem      [][]uint32 `json:"use_item"`
	UseChar      uint32     `json:"use_char"`
	UseCharNum   uint32     `json:"use_char_num"`
	WeaponIDs    []uint32   `json:"weapon_ids"`
	BreakoutView string     `json:"breakout_view"`
}

func GetShipBreakoutConfig(templateID uint32) (*ShipBreakoutConfig, error) {
	return GetShipBreakoutConfigTx(nil, templateID)
}

func GetShipBreakoutConfigTx(_ any, templateID uint32) (*ShipBreakoutConfig, error) {
	entry, err := GetConfigEntry(shipBreakoutCategory, fmt.Sprintf("%d", templateID))
	if err != nil {
		return nil, err
	}
	var config ShipBreakoutConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
