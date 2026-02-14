package orm

import (
	"encoding/json"
	"fmt"
)

const transformDataTemplateCategory = "ShareCfg/transform_data_template.json"

type TransformDataTemplate struct {
	ID          uint32       `json:"id"`
	LevelLimit  uint32       `json:"level_limit"`
	StarLimit   uint32       `json:"star_limit"`
	MaxLevel    uint32       `json:"max_level"`
	UseGold     uint32       `json:"use_gold"`
	UseShip     uint32       `json:"use_ship"`
	UseItem     [][][]uint32 `json:"use_item"`
	ShipID      [][]uint32   `json:"ship_id"`
	EditTrans   []uint32     `json:"edit_trans"`
	SkinID      uint32       `json:"skin_id"`
	SkillID     uint32       `json:"skill_id"`
	ConditionID []uint32     `json:"condition_id"`
}

func GetTransformDataTemplate(id uint32) (*TransformDataTemplate, error) {
	entry, err := GetConfigEntry(transformDataTemplateCategory, fmt.Sprintf("%d", id))
	if err != nil {
		return nil, err
	}
	var config TransformDataTemplate
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (t *TransformDataTemplate) UseItemsForLevel(level uint32) [][]uint32 {
	if level == 0 {
		return nil
	}
	index := int(level - 1)
	if index >= len(t.UseItem) {
		return nil
	}
	return t.UseItem[index]
}
