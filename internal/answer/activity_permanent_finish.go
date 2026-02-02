package answer

import (
	"strconv"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ActivityPermanentFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11208
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11209, err
	}
	activityID := payload.GetActivityId()
	response := protobuf.SC_11209{Result: proto.Uint32(1)}
	if activityID == 0 {
		return client.SendMessage(11209, &response)
	}

	exists, err := configEntryExists("ShareCfg/activity_task_permanent.json", strconv.FormatUint(uint64(activityID), 10))
	if err != nil {
		return 0, 11209, err
	}
	if !exists {
		return client.SendMessage(11209, &response)
	}

	state, err := orm.GetOrCreateActivityPermanentState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11209, err
	}
	if state.CurrentActivityID != activityID {
		return client.SendMessage(11209, &response)
	}

	state.AddFinished(activityID)
	state.CurrentActivityID = 0
	if err := orm.SaveActivityPermanentState(orm.GormDB, state); err != nil {
		return 0, 11209, err
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(11209, &response)
}
