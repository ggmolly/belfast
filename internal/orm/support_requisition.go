package orm

import (
	"encoding/json"
	"fmt"
)

type SupportRequisitionConfig struct {
	Cost          uint32
	RarityWeights []SupportRarityWeight
	MonthlyCap    uint32
}

type SupportRarityWeight struct {
	Rarity uint32
	Weight uint32
}

type supportRequisitionEntry struct {
	KeyValue    uint32          `json:"key_value"`
	Description json.RawMessage `json:"description"`
}

func LoadSupportRequisitionConfig() (SupportRequisitionConfig, error) {
	entry, err := GetConfigEntry("ShareCfg/gameset.json", "supports_config")
	if err != nil {
		return SupportRequisitionConfig{}, err
	}
	var payload supportRequisitionEntry
	if err := json.Unmarshal(entry.Data, &payload); err != nil {
		return SupportRequisitionConfig{}, err
	}
	var parts []json.RawMessage
	if err := json.Unmarshal(payload.Description, &parts); err != nil {
		return SupportRequisitionConfig{}, err
	}
	if len(parts) < 3 {
		return SupportRequisitionConfig{}, fmt.Errorf("supports_config description missing fields")
	}
	var cost uint32
	if err := json.Unmarshal(parts[0], &cost); err != nil {
		return SupportRequisitionConfig{}, err
	}
	var weightsRaw [][]uint32
	if err := json.Unmarshal(parts[1], &weightsRaw); err != nil {
		return SupportRequisitionConfig{}, err
	}
	weights := make([]SupportRarityWeight, 0, len(weightsRaw))
	for _, entry := range weightsRaw {
		if len(entry) != 2 {
			return SupportRequisitionConfig{}, fmt.Errorf("supports_config rarity entry must have 2 values")
		}
		weights = append(weights, SupportRarityWeight{Rarity: entry[0], Weight: entry[1]})
	}
	var cap uint32
	if err := json.Unmarshal(parts[2], &cap); err != nil {
		return SupportRequisitionConfig{}, err
	}
	return SupportRequisitionConfig{Cost: cost, RarityWeights: weights, MonthlyCap: cap}, nil
}
