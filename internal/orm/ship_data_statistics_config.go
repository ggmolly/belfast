package orm

import (
	"encoding/json"
	"fmt"
)

const shipDataStatisticsCategory = "sharecfgdata/ship_data_statistics.json"

type ShipDataStatisticsConfig struct {
	ID     uint32 `json:"id"`
	SkinID uint32 `json:"skin_id"`
}

func GetShipDataStatisticsConfig(templateID uint32) (*ShipDataStatisticsConfig, error) {
	return GetShipDataStatisticsConfigTx(nil, templateID)
}

func GetShipDataStatisticsConfigTx(_ any, templateID uint32) (*ShipDataStatisticsConfig, error) {
	entry, err := GetConfigEntry(shipDataStatisticsCategory, fmt.Sprintf("%d", templateID))
	if err != nil {
		return nil, err
	}
	var config ShipDataStatisticsConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func GetShipBaseSkinIDTx(_ any, templateID uint32) (uint32, error) {
	config, err := GetShipDataStatisticsConfigTx(nil, templateID)
	if err != nil {
		return 0, err
	}
	return config.SkinID, nil
}
