package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
)

const (
	lessonResultOK     = 0
	lessonResultFailed = 1

	lessonItemType         = 10
	lessonItemUsage        = "usage_book"
	tacticsSkillLevelExp   = 100
	defaultSkillMaxLevel   = 10
	skillCancelTypeAuto    = 0
	skillCancelTypeManual  = 1
	itemConfigCategory     = "sharecfgdata/item_data_statistics.json"
	shipTemplateCategory   = "sharecfgdata/ship_data_template.json"
	skillTemplateCategory  = "sharecfgdata/skill_data_template.json"
	skillTemplateCategory2 = "ShareCfg/skill_data_template.json"
)

type lessonItemConfig struct {
	ID       uint32          `json:"id"`
	Type     uint32          `json:"type"`
	Usage    string          `json:"usage"`
	UsageArg json.RawMessage `json:"usage_arg"`
}

type shipSkillConfig struct {
	BuffListDisplay json.RawMessage `json:"buff_list_display"`
}

type skillTemplateConfig struct {
	Type     uint32 `json:"type"`
	MaxLevel uint32 `json:"max_level"`
}

func loadLessonItemConfig(itemID uint32) (*lessonItemConfig, []uint32, error) {
	entry, err := orm.GetConfigEntry(itemConfigCategory, strconv.FormatUint(uint64(itemID), 10))
	if err != nil {
		return nil, nil, err
	}
	var config lessonItemConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, nil, err
	}
	var usageArg []uint32
	if err := decodeUint32Array(config.UsageArg, &usageArg); err != nil {
		return nil, nil, err
	}
	return &config, usageArg, nil
}

func lessonExpFromUsageArg(usageArg []uint32, skillType uint32) (duration uint32, exp uint32, err error) {
	if len(usageArg) < 4 {
		return 0, 0, errors.New("lesson item usage_arg is incomplete")
	}
	duration = usageArg[0]
	baseExp := usageArg[1]
	targetSkillType := usageArg[2]
	bonusPct := usageArg[3]
	if duration == 0 || baseExp == 0 {
		return 0, 0, errors.New("lesson item duration or exp is invalid")
	}
	if targetSkillType != 0 && targetSkillType == skillType {
		exp = uint32(math.Floor(float64(baseExp) * float64(100+bonusPct) / 100.0))
	} else {
		exp = baseExp
	}
	return duration, exp, nil
}

func loadShipSkillByPos(shipTemplateID uint32, skillPos uint32) (uint32, error) {
	entry, err := orm.GetConfigEntry(shipTemplateCategory, strconv.FormatUint(uint64(shipTemplateID), 10))
	if err != nil {
		return 0, err
	}
	var config shipSkillConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return 0, err
	}
	var skillIDs []uint32
	if err := decodeUint32Array(config.BuffListDisplay, &skillIDs); err != nil {
		return 0, err
	}
	if skillPos == 0 || int(skillPos) > len(skillIDs) {
		return 0, db.ErrNotFound
	}
	if skillIDs[skillPos-1] == 0 {
		return 0, db.ErrNotFound
	}
	return skillIDs[skillPos-1], nil
}

func loadSkillTemplate(skillID uint32) (*skillTemplateConfig, error) {
	key := strconv.FormatUint(uint64(skillID), 10)
	entry, err := orm.GetConfigEntry(skillTemplateCategory, key)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			entry, err = orm.GetConfigEntry(skillTemplateCategory2, key)
		}
		if err != nil {
			return nil, err
		}
	}
	var config skillTemplateConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return nil, err
	}
	if config.MaxLevel == 0 {
		config.MaxLevel = defaultSkillMaxLevel
	}
	return &config, nil
}

func decodeUint32Array(raw json.RawMessage, target *[]uint32) error {
	if len(raw) == 0 {
		*target = nil
		return nil
	}
	if err := json.Unmarshal(raw, target); err == nil {
		return nil
	}
	var values []any
	if err := json.Unmarshal(raw, &values); err != nil {
		return err
	}
	result := make([]uint32, 0, len(values))
	for _, value := range values {
		number, ok := value.(float64)
		if !ok {
			return fmt.Errorf("invalid uint32 value in array")
		}
		if number < 0 {
			return fmt.Errorf("invalid uint32 value in array")
		}
		result = append(result, uint32(number))
	}
	*target = result
	return nil
}

func calcGrantedLessonExp(now time.Time, startUnix uint32, finishUnix uint32, totalExp uint32) uint32 {
	if totalExp == 0 {
		return 0
	}
	nowUnix := uint32(now.Unix())
	if nowUnix <= startUnix {
		return 0
	}
	if nowUnix >= finishUnix {
		return totalExp
	}
	duration := finishUnix - startUnix
	if duration == 0 {
		return totalExp
	}
	elapsed := nowUnix - startUnix
	return uint32((uint64(totalExp) * uint64(elapsed)) / uint64(duration))
}

func applyLessonExp(skill *orm.CommanderShipSkill, amount uint32, maxLevel uint32) uint32 {
	if amount == 0 || skill.Level >= maxLevel {
		return 0
	}
	remaining := amount
	granted := uint32(0)
	for remaining > 0 && skill.Level < maxLevel {
		need := tacticsSkillLevelExp - skill.Exp
		if need == 0 {
			skill.Level++
			skill.Exp = 0
			continue
		}
		if remaining < need {
			skill.Exp += remaining
			granted += remaining
			remaining = 0
			break
		}
		skill.Level++
		skill.Exp = 0
		granted += need
		remaining -= need
	}
	if skill.Level >= maxLevel {
		skill.Level = maxLevel
		skill.Exp = 0
	}
	return granted
}
