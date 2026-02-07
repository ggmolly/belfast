package misc

import (
	"encoding/json"
	"errors"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	escortTemplateCategory    = "ShareCfg/escort_template.json"
	escortMapTemplateCategory = "ShareCfg/escort_map_template.json"
)

type EscortTemplate struct {
	ID             uint32          `json:"id"`
	GardroadReward json.RawMessage `json:"gardroad_reward"`
}

type EscortMapTemplate struct {
	ID           uint32          `json:"id"`
	RefreshTime  uint32          `json:"refresh_time"`
	DropByWarn   []uint32        `json:"drop_by_warn"`
	EscortIDList json.RawMessage `json:"escort_id_list"`
}

type EscortConfig struct {
	Templates map[uint32]EscortTemplate
	Maps      []EscortMapTemplate
}

func GetEscortConfig() (*EscortConfig, error) {
	templatesRaw, err := orm.ListConfigEntries(orm.GormDB, escortTemplateCategory)
	if err != nil {
		return nil, err
	}
	mapsRaw, err := orm.ListConfigEntries(orm.GormDB, escortMapTemplateCategory)
	if err != nil {
		return nil, err
	}

	templates := make(map[uint32]EscortTemplate, len(templatesRaw))
	for _, entry := range templatesRaw {
		var template EscortTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return nil, err
		}
		templates[template.ID] = template
	}

	maps := make([]EscortMapTemplate, 0, len(mapsRaw))
	for _, entry := range mapsRaw {
		var template EscortMapTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return nil, err
		}
		maps = append(maps, template)
	}

	return &EscortConfig{
		Templates: templates,
		Maps:      maps,
	}, nil
}

type escortPosState struct {
	MapID     uint32 `json:"map_id"`
	ChapterID uint32 `json:"chapter_id"`
}

func LoadEscortState(accountID uint32) ([]*protobuf.ESCORT_INFO, error) {
	var states []orm.EscortState
	if err := orm.GormDB.Where("account_id = ?", accountID).Order("line_id asc").Find(&states).Error; err != nil {
		return nil, err
	}
	infos := make([]*protobuf.ESCORT_INFO, 0, len(states))
	for _, state := range states {
		positions := []*protobuf.ESCORT_POS{}
		if len(state.MapPositions) != 0 {
			var raw []escortPosState
			if err := json.Unmarshal(state.MapPositions, &raw); err != nil {
				return nil, err
			}
			positions = make([]*protobuf.ESCORT_POS, 0, len(raw))
			for _, pos := range raw {
				positions = append(positions, &protobuf.ESCORT_POS{
					MapId:     proto.Uint32(pos.MapID),
					ChapterId: proto.Uint32(pos.ChapterID),
				})
			}
		}
		infos = append(infos, &protobuf.ESCORT_INFO{
			LineId:         proto.Uint32(state.LineID),
			AwardTimestamp: proto.Uint32(state.AwardTimestamp),
			FlashTimestamp: proto.Uint32(state.FlashTimestamp),
			Map:            positions,
		})
	}
	return infos, nil
}

func UpdateEscortTimestamps(accountID uint32, lineID uint32, awardTS uint32, flashTS uint32) error {
	var state orm.EscortState
	err := orm.GormDB.Where("account_id = ? AND line_id = ?", accountID, lineID).First(&state).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		state = orm.EscortState{
			AccountID:    accountID,
			LineID:       lineID,
			MapPositions: json.RawMessage("[]"),
		}
		state.AwardTimestamp = awardTS
		state.FlashTimestamp = flashTS
		return orm.GormDB.Create(&state).Error
	}
	state.AwardTimestamp = awardTS
	state.FlashTimestamp = flashTS
	return orm.GormDB.Save(&state).Error
}
