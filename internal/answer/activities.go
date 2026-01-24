package answer

import (
	"encoding/json"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func Activities(buffer *[]byte, client *connection.Client) (int, int, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/activity_template.json")
	if err != nil {
		return 0, 11200, err
	}
	response := protobuf.SC_11200{
		ActivityList: make([]*protobuf.ACTIVITYINFO, 0, len(entries)),
	}
	for _, entry := range entries {
		var template activityTemplate
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 11200, err
		}
		info, err := buildActivityInfo(template, activityStopTime(template.Time))
		if err != nil {
			return 0, 11200, err
		}
		if info == nil {
			continue
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
