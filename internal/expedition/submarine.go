package expedition

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/orm"
)

const (
	submarineDailyTemplateCategory = "ShareCfg/expedition_daily_template.json"
	submarineDataTemplateCategory  = "ShareCfg/expedition_data_template.json"

	submarineDailyExpeditionID = uint32(501)
	submarineStageType         = uint32(15)
)

type SubmarineChapter struct {
	ChapterID uint32
	MinLevel  uint32
	Index     uint32
}

type submarineDailyTemplate struct {
	ID                      uint32     `json:"id"`
	ExpeditionLvLimitList   [][]uint32 `json:"expedition_and_lv_limit_list"`
	ExpeditionAndLvLimitRaw json.RawMessage
}

type submarineDataTemplate struct {
	ID   uint32 `json:"id"`
	Type uint32 `json:"type"`
}

func LoadSubmarineChapters() ([]SubmarineChapter, error) {
	limits, err := loadSubmarineLevelLimits()
	if err != nil {
		return nil, err
	}
	entries, err := orm.ListConfigEntries(orm.GormDB, submarineDataTemplateCategory)
	if err != nil {
		return nil, err
	}
	chapters := make([]SubmarineChapter, 0)
	for _, entry := range entries {
		var template submarineDataTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return nil, err
		}
		if template.Type != submarineStageType {
			continue
		}
		if template.ID < 1000 || template.ID > 1005 {
			continue
		}
		minLevel, ok := limits[template.ID]
		if !ok {
			continue
		}
		chapters = append(chapters, SubmarineChapter{
			ChapterID: template.ID,
			MinLevel:  minLevel,
			Index:     template.ID - 1000,
		})
	}
	return chapters, nil
}

func loadSubmarineLevelLimits() (map[uint32]uint32, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, submarineDailyTemplateCategory)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Key != "501" {
			// fall back to parsing id for older seeds
			var peek struct {
				ID uint32 `json:"id"`
			}
			if err := json.Unmarshal(entry.Data, &peek); err == nil {
				if peek.ID != submarineDailyExpeditionID {
					continue
				}
			} else {
				continue
			}
		}

		var template struct {
			ID                    uint32     `json:"id"`
			ExpeditionLvLimitList [][]uint32 `json:"expedition_and_lv_limit_list"`
		}
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return nil, err
		}
		limits := make(map[uint32]uint32)
		for _, pair := range template.ExpeditionLvLimitList {
			if len(pair) < 2 {
				continue
			}
			limits[pair[0]] = pair[1]
		}
		return limits, nil
	}
	return map[uint32]uint32{}, nil
}
