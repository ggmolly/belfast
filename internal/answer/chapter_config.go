package answer

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ggmolly/belfast/internal/orm"
	"gorm.io/gorm"
)

const (
	chapterTemplateCategory     = "sharecfgdata/chapter_template.json"
	chapterTemplateLoopCategory = "sharecfgdata/chapter_template_loop.json"
	itemDataStatsCategory       = "sharecfgdata/item_data_statistics.json"
	benefitBuffCategory         = "ShareCfg/benefit_buff_template.json"
)

type chapterTemplate struct {
	ID                 uint32     `json:"id"`
	Grids              [][]any    `json:"grids"`
	AmmoTotal          uint32     `json:"ammo_total"`
	AmmoSubmarine      uint32     `json:"ammo_submarine"`
	GroupNum           uint32     `json:"group_num"`
	SubmarineNum       uint32     `json:"submarine_num"`
	SupportGroupNum    uint32     `json:"support_group_num"`
	IsAmbush           uint32     `json:"is_ambush"`
	InvestigationRatio uint32     `json:"investigation_ratio"`
	AvoidRatio         uint32     `json:"avoid_ratio"`
	AmbushRatioExtra   [][]int32  `json:"ambush_ratio_extra"`
	ChapterStrategy    []uint32   `json:"chapter_strategy"`
	BossExpeditionID   []uint32   `json:"boss_expedition_id"`
	ExpeditionWeight   [][]any    `json:"expedition_id_weight_list"`
	EliteExpeditions   []uint32   `json:"elite_expedition_list"`
	AmbushExpeditions  []uint32   `json:"ambush_expedition_list"`
	GuarderExpeditions []uint32   `json:"guarder_expedition_list"`
	Awards             [][]uint32 `json:"awards"`
	StarRequire1       uint32     `json:"star_require_1"`
	StarRequire2       uint32     `json:"star_require_2"`
	StarRequire3       uint32     `json:"star_require_3"`
	Num1               uint32     `json:"num_1"`
	Num2               uint32     `json:"num_2"`
	Num3               uint32     `json:"num_3"`
	ProgressBoss       uint32     `json:"progress_boss"`
	Oil                uint32     `json:"oil"`
	Time               uint32     `json:"time"`
}

type itemDataStatisticsEntry struct {
	ID       uint32          `json:"id"`
	UsageArg json.RawMessage `json:"usage_arg"`
}

type benefitBuffEntry struct {
	ID               uint32 `json:"id"`
	BenefitType      string `json:"benefit_type"`
	BenefitEffect    string `json:"benefit_effect"`
	BenefitCondition string `json:"benefit_condition"`
}

func loadChapterTemplate(chapterID uint32, loopFlag uint32) (*chapterTemplate, error) {
	category := chapterTemplateCategory
	if loopFlag != 0 {
		category = chapterTemplateLoopCategory
	}
	entry, err := orm.GetConfigEntry(orm.GormDB, category, fmt.Sprintf("%d", chapterID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var template chapterTemplate
	if err := json.Unmarshal(entry.Data, &template); err != nil {
		return nil, err
	}
	return &template, nil
}

func loadItemUsageArg(itemID uint32) ([]uint32, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, itemDataStatsCategory, fmt.Sprintf("%d", itemID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var stats itemDataStatisticsEntry
	if err := json.Unmarshal(entry.Data, &stats); err != nil {
		return nil, err
	}
	return decodeUsageArgUint32(stats.UsageArg)
}

func calculateOperationItemCostRate(itemID uint32) (float64, error) {
	if itemID == 0 {
		return 1, nil
	}
	ids, err := loadItemUsageArg(itemID)
	if err != nil {
		return 0, err
	}
	rate := 1.0
	for _, buffID := range ids {
		entry, err := loadBenefitBuff(buffID)
		if err != nil {
			return 0, err
		}
		if entry == nil || entry.BenefitType != "more_oil" {
			continue
		}
		effect, err := strconv.ParseFloat(entry.BenefitEffect, 64)
		if err != nil {
			continue
		}
		rate += effect * 0.01
	}
	return math.Max(1, rate), nil
}

func findOperationBuffID(itemID uint32) (uint32, error) {
	if itemID == 0 {
		return 0, nil
	}
	entries, err := orm.ListConfigEntries(orm.GormDB, benefitBuffCategory)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		var buff benefitBuffEntry
		if err := json.Unmarshal(entry.Data, &buff); err != nil {
			return 0, err
		}
		if buff.BenefitType != "desc" {
			continue
		}
		condition, err := strconv.ParseUint(buff.BenefitCondition, 10, 32)
		if err != nil {
			continue
		}
		if uint32(condition) == itemID {
			return buff.ID, nil
		}
	}
	return 0, nil
}

func loadBenefitBuff(buffID uint32) (*benefitBuffEntry, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, benefitBuffCategory, fmt.Sprintf("%d", buffID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var buff benefitBuffEntry
	if err := json.Unmarshal(entry.Data, &buff); err != nil {
		return nil, err
	}
	return &buff, nil
}

func decodeUsageArgUint32(raw json.RawMessage) ([]uint32, error) {
	normalized, err := normalizeChapterUsageArg(raw)
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	var ids []uint32
	if err := json.Unmarshal(normalized, &ids); err == nil {
		return ids, nil
	}
	var generic []any
	if err := json.Unmarshal(normalized, &generic); err != nil {
		return nil, err
	}
	ids = make([]uint32, 0, len(generic))
	for _, value := range generic {
		switch typed := value.(type) {
		case float64:
			ids = append(ids, uint32(typed))
		case string:
			parsed, err := strconv.ParseUint(typed, 10, 32)
			if err != nil {
				continue
			}
			ids = append(ids, uint32(parsed))
		}
	}
	return ids, nil
}

func normalizeChapterUsageArg(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		text = strings.TrimSpace(text)
		if text == "" {
			text = "[]"
		}
		if !json.Valid([]byte(text)) {
			return nil, fmt.Errorf("invalid usage_arg: %s", text)
		}
		return json.RawMessage([]byte(text)), nil
	}
	return raw, nil
}
