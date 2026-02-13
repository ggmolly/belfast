package orm

import (
	"encoding/json"
	"fmt"
)

const spweaponDataStatisticsCategory = "ShareCfg/spweapon_data_statistics.json"

type SpWeaponDataStatisticsConfig struct {
	ID           uint32 `json:"id"`
	UpgradeID    uint32 `json:"upgrade_id"`
	Value1Random uint32 `json:"value_1_random"`
	Value2Random uint32 `json:"value_2_random"`
}

func GetSpWeaponDataStatisticsConfigTx(templateID uint32) (*SpWeaponDataStatisticsConfig, error) {
	entry, err := GetConfigEntry(spweaponDataStatisticsCategory, fmt.Sprintf("%d", templateID))
	if err != nil {
		return nil, err
	}
	var cfg SpWeaponDataStatisticsConfig
	if err := json.Unmarshal(entry.Data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
