package orm

import (
	"encoding/json"
	"fmt"
)

type ShipTemplateConfig struct {
	ID           uint32 `json:"id"`
	StrengthenID uint32 `json:"strengthen_id"`
	GroupType    uint32 `json:"group_type"`
	MaxLevel     uint32 `json:"max_level"`
}

func GetShipTemplateConfig(templateID uint32) (*ShipTemplateConfig, error) {
	return GetShipTemplateConfigTx(nil, templateID)
}

func GetShipTemplateConfigTx(_ any, templateID uint32) (*ShipTemplateConfig, error) {
	entry, err := GetConfigEntry(shipDataTemplateCategory, fmt.Sprintf("%d", templateID))
	if err != nil {
		return nil, err
	}
	var config ShipTemplateConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
