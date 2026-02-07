package orm

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const composeDataTemplateCategory = "ShareCfg/compose_data_template.json"

type ComposeDataTemplateEntry struct {
	ID          uint32 `json:"id"`
	EquipID     uint32 `json:"equip_id"`
	MaterialID  uint32 `json:"material_id"`
	MaterialNum uint32 `json:"material_num"`
	GoldNum     uint32 `json:"gold_num"`
}

func GetComposeDataTemplateEntry(db *gorm.DB, id uint32) (*ComposeDataTemplateEntry, error) {
	entry, err := GetConfigEntry(db, composeDataTemplateCategory, fmt.Sprintf("%d", id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var parsed ComposeDataTemplateEntry
	if err := json.Unmarshal(entry.Data, &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}
