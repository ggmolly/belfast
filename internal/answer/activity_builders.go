package answer

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type townLevelConfig struct {
	ID          uint32     `json:"id"`
	UnlockChara uint32     `json:"unlock_chara"`
	UnlockWork  [][]uint32 `json:"unlock_work"`
}

func buildActivityInfo(template activityTemplate, stopTime uint32) (*protobuf.ACTIVITYINFO, error) {
	if template.Type == activityTypePuzzle {
		ok, err := validatePuzzleActivity(template.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
	}
	if template.Type == activityTypeNewServerTask {
		ok, err := validateNewServerTaskActivity(template.ConfigData)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil
		}
	}
	info := baseActivityInfo(template, stopTime)
	if template.Type == activityTypeTown {
		return buildTownActivityInfo(info)
	}
	if template.Type == activityTypeBossBattleMark2 {
		return buildBossBattleMark2ActivityInfo(info), nil
	}
	return info, nil
}

func baseActivityInfo(template activityTemplate, stopTime uint32) *protobuf.ACTIVITYINFO {
	return &protobuf.ACTIVITYINFO{
		Id:                proto.Uint32(template.ID),
		StopTime:          proto.Uint32(stopTime),
		Data1List:         []uint32{},
		Data2List:         []uint32{},
		Data3List:         []uint32{},
		Data4List:         []uint32{},
		Date1KeyValueList: []*protobuf.KEYVALUELIST_P11{},
		GroupList:         []*protobuf.GROUPINFO_P11{},
		CollectionList:    []*protobuf.COLLECTIONINFO{},
		TaskList:          []*protobuf.TASKINFO{},
		BuffList:          []*protobuf.BENEFITBUFF{},
	}
}

func buildTownActivityInfo(info *protobuf.ACTIVITYINFO) (*protobuf.ACTIVITYINFO, error) {
	const townLevel = uint32(1)

	level, err := loadTownLevelConfig(townLevel)
	if err != nil {
		return nil, err
	}

	startTime := uint32(time.Now().Unix())
	workplaceCapacity := 0
	if len(level.UnlockWork) > 0 {
		workplaceCapacity = len(level.UnlockWork[0])
	}
	workplaces := make([]*protobuf.KEYVALUE_P11, 0, workplaceCapacity)
	if len(level.UnlockWork) > 0 {
		for _, workplaceID := range level.UnlockWork[0] {
			workplaces = append(workplaces, &protobuf.KEYVALUE_P11{
				Key:   proto.Uint32(workplaceID),
				Value: proto.Uint32(startTime),
			})
		}
	}

	info.Data1 = proto.Uint32(0)
	info.Data2 = proto.Uint32(townLevel)
	info.Date1KeyValueList = append(info.Date1KeyValueList, &protobuf.KEYVALUELIST_P11{
		Key:       proto.Uint32(1),
		ValueList: workplaces,
	})
	return info, nil
}

func buildBossBattleMark2ActivityInfo(info *protobuf.ACTIVITYINFO) *protobuf.ACTIVITYINFO {
	info.Date1KeyValueList = append(info.Date1KeyValueList,
		&protobuf.KEYVALUELIST_P11{
			Key:       proto.Uint32(1),
			ValueList: []*protobuf.KEYVALUE_P11{},
		},
		&protobuf.KEYVALUELIST_P11{
			Key:       proto.Uint32(2),
			ValueList: []*protobuf.KEYVALUE_P11{},
		},
	)
	return info
}

func loadTownLevelConfig(level uint32) (townLevelConfig, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/activity_town_level.json", strconv.FormatUint(uint64(level), 10))
	if err != nil {
		return townLevelConfig{}, err
	}
	var config townLevelConfig
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return townLevelConfig{}, err
	}
	return config, nil
}

func validatePuzzleActivity(activityID uint32) (bool, error) {
	_, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/activity_event_picturepuzzle.json", strconv.FormatUint(uint64(activityID), 10))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// TODO: Align puzzle activity filtering with region-specific event schedules.
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func validateNewServerTaskActivity(configData json.RawMessage) (bool, error) {
	if len(configData) == 0 {
		// TODO: Confirm whether empty config_data should be treated as inactive.
		return false, nil
	}
	var taskGroups [][]uint32
	if err := json.Unmarshal(configData, &taskGroups); err != nil {
		return false, err
	}
	for _, group := range taskGroups {
		for _, taskID := range group {
			_, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/task_data_template.json", strconv.FormatUint(uint64(taskID), 10))
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// TODO: Remove filtering once new server tasks are backed by persisted task data.
					return false, nil
				}
				return false, err
			}
		}
	}
	return true, nil
}
