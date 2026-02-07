package orm

import (
	"encoding/json"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

const spweaponUpgradeCategory = "ShareCfg/spweapon_upgrade.json"

type SpWeaponResetMaterial struct {
	ItemID uint32
	Count  uint32
}

type SpWeaponUpgradeConfig struct {
	ID           uint32
	ResetUseItem []SpWeaponResetMaterial
}

func GetSpWeaponUpgradeConfigTx(db *gorm.DB, upgradeID uint32) (*SpWeaponUpgradeConfig, error) {
	entry, err := GetConfigEntry(db, spweaponUpgradeCategory, strconv.FormatUint(uint64(upgradeID), 10))
	if err != nil {
		return nil, err
	}

	var raw struct {
		ID           uint32          `json:"id"`
		ResetUseItem json.RawMessage `json:"reset_use_item"`
	}
	if err := json.Unmarshal(entry.Data, &raw); err != nil {
		return nil, err
	}
	materials, err := parseSpWeaponResetMaterials(raw.ResetUseItem)
	if err != nil {
		return nil, err
	}
	return &SpWeaponUpgradeConfig{ID: raw.ID, ResetUseItem: materials}, nil
}

func parseSpWeaponResetMaterials(raw json.RawMessage) ([]SpWeaponResetMaterial, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var pairs [][]uint32
	if err := json.Unmarshal(raw, &pairs); err != nil {
		return nil, err
	}
	result := make([]SpWeaponResetMaterial, 0, len(pairs))
	for _, pair := range pairs {
		if len(pair) != 2 {
			return nil, errors.New("invalid spweapon reset_use_item")
		}
		result = append(result, SpWeaponResetMaterial{ItemID: pair[0], Count: pair[1]})
	}
	return result, nil
}
