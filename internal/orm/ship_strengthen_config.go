package orm

import (
	"encoding/json"
	"fmt"
)

const shipStrengthenCategory = "ShareCfg/ship_data_strengthen.json"

type ShipStrengthenConfig struct {
	ID         uint32   `json:"id"`
	AttrExp    []uint32 `json:"attr_exp"`
	Durability []uint32 `json:"durability"`
	LevelExp   []uint32 `json:"level_exp"`
}

func GetShipStrengthenConfig(templateID uint32) (*ShipStrengthenConfig, error) {
	return GetShipStrengthenConfigTx(nil, templateID)
}

func GetShipStrengthenConfigTx(_ any, templateID uint32) (*ShipStrengthenConfig, error) {
	entry, err := GetConfigEntry(shipStrengthenCategory, fmt.Sprintf("%d", templateID))
	if err != nil {
		return nil, err
	}
	var config ShipStrengthenConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
