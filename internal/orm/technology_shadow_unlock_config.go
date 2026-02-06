package orm

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

const technologyShadowUnlockCategory = "ShareCfg/technology_shadow_unlock.json"

type TechnologyShadowUnlockConfig struct {
	ID        uint32 `json:"id"`
	Type      uint32 `json:"type"`
	TargetNum uint32 `json:"target_num"`
}

func GetTechnologyShadowUnlockConfig(id uint32) (*TechnologyShadowUnlockConfig, error) {
	return GetTechnologyShadowUnlockConfigTx(GormDB, id)
}

func GetTechnologyShadowUnlockConfigTx(db *gorm.DB, id uint32) (*TechnologyShadowUnlockConfig, error) {
	entry, err := GetConfigEntry(db, technologyShadowUnlockCategory, fmt.Sprintf("%d", id))
	if err != nil {
		return nil, err
	}
	var config TechnologyShadowUnlockConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
