package answer

import (
	"encoding/json"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func Activities(buffer *[]byte, client *connection.Client) (int, int, error) {
	allowlist, err := loadActivityAllowlist()
	if err != nil {
		return 0, 11200, err
	}
	state, err := orm.GetOrCreateActivityPermanentState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 11200, err
	}
	permanentIDs, err := loadPermanentActivityIDSet()
	if err != nil {
		return 0, 11200, err
	}
	finished := filterPermanentActivityIDs(orm.ToUint32List(state.FinishedActivityIDs), permanentIDs)
	finishedSet := make(map[uint32]struct{}, len(finished))
	for _, id := range finished {
		finishedSet[id] = struct{}{}
	}

	activityIDs := make([]uint32, 0, len(allowlist)+1)
	for _, activityID := range allowlist {
		if _, ok := finishedSet[activityID]; ok {
			continue
		}
		activityIDs = append(activityIDs, activityID)
	}
	if state.CurrentActivityID != 0 {
		if _, ok := permanentIDs[state.CurrentActivityID]; ok {
			if _, finished := finishedSet[state.CurrentActivityID]; !finished {
				activityIDs = appendUniqueUint32(activityIDs, state.CurrentActivityID)
			}
		}
	}
	response := protobuf.SC_11200{
		ActivityList: make([]*protobuf.ACTIVITYINFO, 0, len(activityIDs)),
	}
	if len(activityIDs) == 0 {
		// TODO: Allow a config toggle to fall back to ShareCfg activity defaults.
		return client.SendMessage(11200, &response)
	}
	for _, activityID := range activityIDs {
		template, err := loadActivityTemplate(activityID)
		if err != nil {
			return 0, 11200, err
		}
		info, err := buildActivityInfo(template, activityStopTime(template.Time))
		if err != nil {
			return 0, 11200, err
		}
		if info == nil {
			continue
		}
		groups, found, err := orm.LoadActivityFleetGroups(client.Commander.CommanderID, template.ID)
		if err != nil {
			return 0, 11200, err
		}
		if found {
			info.GroupList = activityFleetGroupsToProto(groups)
		}
		response.ActivityList = append(response.ActivityList, info)
	}
	return client.SendMessage(11200, &response)
}

func activityStopTime(raw json.RawMessage) uint32 {
	var label string
	if err := json.Unmarshal(raw, &label); err == nil {
		return 0
	}
	var value []any
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0
	}
	if len(value) < 3 {
		return 0
	}
	typeTag, ok := value[0].(string)
	if !ok || typeTag != "timer" {
		return 0
	}
	end, ok := value[2].([]any)
	if !ok || len(end) != 2 {
		return 0
	}
	date, ok := end[0].([]any)
	if !ok || len(date) != 3 {
		return 0
	}
	clock, ok := end[1].([]any)
	if !ok || len(clock) != 3 {
		return 0
	}
	year, ok := parseJSONInt(date[0])
	if !ok {
		return 0
	}
	month, ok := parseJSONInt(date[1])
	if !ok {
		return 0
	}
	day, ok := parseJSONInt(date[2])
	if !ok {
		return 0
	}
	hour, ok := parseJSONInt(clock[0])
	if !ok {
		return 0
	}
	minute, ok := parseJSONInt(clock[1])
	if !ok {
		return 0
	}
	second, ok := parseJSONInt(clock[2])
	if !ok {
		return 0
	}
	stop := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return uint32(stop.Unix())
}

func parseJSONInt(value any) (int, bool) {
	number, ok := value.(float64)
	if !ok {
		return 0, false
	}
	return int(number), true
}

func loadActivityAllowlist() ([]uint32, error) {
	var entry orm.ConfigEntry
	result := orm.GormDB.Where("category = ? AND key = ?", "ServerCfg/activities.json", "allowlist").Limit(1).Find(&entry)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return []uint32{}, nil
	}
	var allowlist []uint32
	if err := json.Unmarshal(entry.Data, &allowlist); err != nil {
		return nil, err
	}
	return allowlist, nil
}
