package answer

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type townLevelConfig struct {
	ID          uint32     `json:"id"`
	UnlockChara uint32     `json:"unlock_chara"`
	UnlockWork  [][]uint32 `json:"unlock_work"`
}

func buildActivityInfo(template activityTemplate, stopTime uint32) (*protobuf.ACTIVITYINFO, error) {
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
