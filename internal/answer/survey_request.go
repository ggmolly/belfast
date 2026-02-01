package answer

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

var errSurveyActivityAmbiguous = errors.New("multiple survey activities for survey id")

func SurveyRequest(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11025
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11026, err
	}
	response := protobuf.SC_11026{
		Result: proto.Uint32(0),
	}
	template, err := findSurveyActivityTemplate(payload.GetSurveyId())
	if err != nil {
		return 0, 11026, err
	}
	if template == nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11026, &response)
	}
	open, err := isSurveyActivityOpen(*template, client.Commander.Level)
	if err != nil {
		return 0, 11026, err
	}
	if !open {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11026, &response)
	}
	if err := orm.SetCommanderSurveyCompleted(orm.GormDB, client.Commander.CommanderID, payload.GetSurveyId(), time.Now().UTC()); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11026, &response)
}

func findSurveyActivityTemplate(surveyID uint32) (*activityTemplate, error) {
	allowlist, err := loadActivityAllowlist()
	if err != nil {
		return nil, err
	}
	var match *activityTemplate
	for _, activityID := range allowlist {
		template, err := loadActivityTemplate(activityID)
		if err != nil {
			return nil, err
		}
		if template.Type != activityTypeSurvey || template.ConfigID != surveyID {
			continue
		}
		if match != nil {
			return nil, errSurveyActivityAmbiguous
		}
		copy := template
		match = &copy
	}
	return match, nil
}

func isSurveyActivityOpen(template activityTemplate, commanderLevel int) (bool, error) {
	enabled, minLevel, err := parseSurveyConfigData(template.ConfigData)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}
	if commanderLevel < int(minLevel) {
		return false, nil
	}
	stopTime := activityStopTime(template.Time)
	if stopTime != 0 && time.Now().UTC().Unix() > int64(stopTime) {
		return false, nil
	}
	return true, nil
}

func parseSurveyConfigData(config json.RawMessage) (bool, uint32, error) {
	if len(config) == 0 {
		return false, 0, nil
	}
	var values []uint32
	if err := json.Unmarshal(config, &values); err == nil {
		if len(values) < 2 {
			return false, 0, nil
		}
		return values[0] == 1, values[1], nil
	}
	var raw []any
	if err := json.Unmarshal(config, &raw); err != nil {
		return false, 0, err
	}
	if len(raw) < 2 {
		return false, 0, nil
	}
	flag, ok := parseJSONUint(raw[0])
	if !ok {
		return false, 0, errors.New("unsupported survey config flag")
	}
	level, ok := parseJSONUint(raw[1])
	if !ok {
		return false, 0, errors.New("unsupported survey config level")
	}
	return flag == 1, level, nil
}
