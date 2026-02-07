package answer

import (
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/ggmolly/belfast/internal/config"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	shipEvaluationCommentMaxRunes = 80
	shipEvaluationDailyCommentMax = 5
	shipEvaluationCommentMinLevel = 1
)

type shipDiscussState struct {
	mu sync.Mutex

	shipGroupID uint32
	dayKey      string

	nextDiscussID     uint32
	discussCount      uint32
	dailyDiscussCount uint32
	discussList       []*protobuf.DISCUSS_INFO

	// Reset on day rollover to bound memory growth.
	reviewedDiscussByCommander map[uint32]map[uint32]struct{}
}

var (
	shipDiscussStoreMu sync.Mutex
	shipDiscussStore   = map[uint32]*shipDiscussState{}
)

func getShipDiscussState(shipGroupID uint32, now time.Time) *shipDiscussState {
	dayKey := now.UTC().Format("2006-01-02")

	shipDiscussStoreMu.Lock()
	state, ok := shipDiscussStore[shipGroupID]
	if !ok {
		state = &shipDiscussState{shipGroupID: shipGroupID, dayKey: dayKey}
		shipDiscussStore[shipGroupID] = state
	}
	shipDiscussStoreMu.Unlock()

	state.mu.Lock()
	if state.dayKey != dayKey {
		state.dayKey = dayKey
		state.dailyDiscussCount = 0
		state.reviewedDiscussByCommander = nil
	}
	state.mu.Unlock()

	return state
}

func commentContainsBannedWord(comment string) (bool, error) {
	createConfig := config.Current().CreatePlayer
	if len(createConfig.NameBlacklist) > 0 {
		lower := strings.ToLower(comment)
		for _, blocked := range createConfig.NameBlacklist {
			blocked = strings.TrimSpace(blocked)
			if blocked == "" {
				continue
			}
			if strings.Contains(lower, strings.ToLower(blocked)) {
				return true, nil
			}
		}
	}
	if createConfig.NameIllegalPattern != "" {
		matcher, err := regexp.Compile(createConfig.NameIllegalPattern)
		if err != nil {
			return false, err
		}
		if matcher.MatchString(comment) {
			return true, nil
		}
	}
	return false, nil
}

func countShipHearts(shipGroupID uint32) (uint32, error) {
	var count int64
	if err := orm.GormDB.Model(&orm.Like{}).Where("group_id = ?", shipGroupID).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint32(count), nil
}

func PostShipEvaluationComment(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17103
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17104, err
	}

	response := protobuf.SC_17104{
		Result:    proto.Uint32(0),
		NeedLevel: proto.Uint32(0),
	}
	if client.Commander == nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17104, &response)
	}

	comment := payload.GetContext()
	if strings.TrimSpace(comment) == "" {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17104, &response)
	}
	if utf8.RuneCountInString(comment) > shipEvaluationCommentMaxRunes {
		response.Result = proto.Uint32(2011)
		return client.SendMessage(17104, &response)
	}

	blocked, err := commentContainsBannedWord(comment)
	if err != nil {
		return 0, 17104, err
	}
	if blocked {
		response.Result = proto.Uint32(2013)
		return client.SendMessage(17104, &response)
	}

	shipGroupID := payload.GetShipGroupId()
	if shipGroupID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17104, &response)
	}

	if client.Commander.Level < shipEvaluationCommentMinLevel {
		response.Result = proto.Uint32(41)
		response.NeedLevel = proto.Uint32(shipEvaluationCommentMinLevel)
		return client.SendMessage(17104, &response)
	}

	now := time.Now()
	state := getShipDiscussState(shipGroupID, now)
	heartCount, err := countShipHearts(shipGroupID)
	if err != nil {
		return 0, 17104, err
	}

	state.mu.Lock()
	if state.dailyDiscussCount >= shipEvaluationDailyCommentMax {
		state.mu.Unlock()
		response.Result = proto.Uint32(1)
		return client.SendMessage(17104, &response)
	}

	state.nextDiscussID++
	entry := &protobuf.DISCUSS_INFO{
		Id:        proto.Uint32(state.nextDiscussID),
		NickName:  proto.String(client.Commander.Name),
		Context:   proto.String(comment),
		GoodCount: proto.Uint32(0),
		BadCount:  proto.Uint32(0),
	}

	state.discussList = append(state.discussList, entry)
	state.discussCount++
	state.dailyDiscussCount++

	discussListCopy := make([]*protobuf.DISCUSS_INFO, len(state.discussList))
	copy(discussListCopy, state.discussList)
	shipDiscuss := &protobuf.SHIP_DISCUSS_INFO{
		ShipGroupId:       proto.Uint32(shipGroupID),
		DiscussCount:      proto.Uint32(state.discussCount),
		HeartCount:        proto.Uint32(heartCount),
		DiscussList:       discussListCopy,
		DailyDiscussCount: proto.Uint32(state.dailyDiscussCount),
	}
	state.mu.Unlock()

	response.ShipDiscuss = shipDiscuss
	return client.SendMessage(17104, &response)
}
