package answer

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type activityEventChallenge struct {
	ID            uint32       `json:"id"`
	Buff          []uint32     `json:"buff"`
	InfiniteStage [][][]uint32 `json:"infinite_stage"`
}

func ChallengeInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_24004
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 24005, err
	}

	activity, err := loadActivityTemplate(payload.GetActivityId())
	if err != nil {
		return 0, 24005, err
	}
	if activity.Type != activityTypeChallenge {
		return 0, 24005, fmt.Errorf("unexpected challenge activity type: %d", activity.Type)
	}

	config, err := loadActivityEventChallenge(activity)
	if err != nil {
		return 0, 24005, err
	}

	seasonID := uint32(1)
	currentChallenge := &protobuf.CHALLENGEINFO{
		SeasonMaxScore:   proto.Uint32(0),
		ActivityMaxScore: proto.Uint32(0),
		SeasonMaxLevel:   proto.Uint32(0),
		ActivityMaxLevel: proto.Uint32(0),
		SeasonId:         proto.Uint32(seasonID),
		DungeonIdList:    challengeDungeonList(config, seasonID),
		BuffList:         config.Buff,
	}

	response := protobuf.SC_24005{
		Result:           proto.Uint32(0),
		CurrentChallenge: currentChallenge,
		// TODO: Populate user challenge data once challenge progress is persisted.
		UserChallenge: []*protobuf.USERCHALLENGEINFO{},
	}
	return client.SendMessage(24005, &response)
}

func loadActivityEventChallenge(activity activityTemplate) (activityEventChallenge, error) {
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/activity_event_challenge.json", strconv.FormatUint(uint64(activity.ConfigID), 10))
	if err != nil {
		return activityEventChallenge{}, err
	}
	var config activityEventChallenge
	if err := json.Unmarshal(entry.Data, &config); err != nil {
		return activityEventChallenge{}, err
	}
	return config, nil
}

func challengeDungeonList(config activityEventChallenge, seasonID uint32) []uint32 {
	// TODO: Select dungeon list based on actual season/rotation rules.
	seasonIndex := int(seasonID - 1)
	return config.InfiniteStage[seasonIndex][0]
}
