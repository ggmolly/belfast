package answer

import (
	"strconv"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ActivityPermanentStart(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11206
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11207, err
	}
	activityID := payload.GetActivityId()
	response := protobuf.SC_11207{Result: proto.Uint32(1)}
	if activityID == 0 {
		return client.SendMessage(11207, &response)
	}

	exists, err := configEntryExists("ShareCfg/activity_task_permanent.json", strconv.FormatUint(uint64(activityID), 10))
	if err != nil {
		return 0, 11207, err
	}
	if !exists {
		return client.SendMessage(11207, &response)
	}

	state, err := orm.GetOrCreateActivityPermanentState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11207, err
	}
	if state.HasFinished(activityID) {
		return client.SendMessage(11207, &response)
	}
	if state.CurrentActivityID != 0 && state.CurrentActivityID != activityID {
		return client.SendMessage(11207, &response)
	}

	if state.CurrentActivityID != activityID {
		state.CurrentActivityID = activityID
		if err := orm.SaveActivityPermanentState(orm.GormDB, state); err != nil {
			return 0, 11207, err
		}
	}

	template, err := loadActivityTemplate(activityID)
	if err != nil {
		return 0, 11207, err
	}
	info, err := buildActivityInfo(template, activityStopTime(template.Time))
	if err != nil {
		return 0, 11207, err
	}
	if info == nil {
		return client.SendMessage(11207, &response)
	}
	update := protobuf.SC_11201{ActivityInfo: info}
	if _, _, err := client.SendMessage(11201, &update); err != nil {
		return 0, 11207, err
	}

	response.Result = proto.Uint32(0)
	return client.SendMessage(11207, &response)
}
