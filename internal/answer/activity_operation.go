package answer

import (
	"encoding/json"
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type singleEventConfig struct {
	ID   uint32 `json:"id"`
	Type uint32 `json:"type"`
}

func ActivityOperation(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11202
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11203, err
	}

	template, err := loadActivityTemplate(payload.GetActivityId())
	if err != nil {
		return 0, 11203, err
	}

	switch template.Type {
	case activityTypeEventSingle:
		if payload.GetCmd() != activityCmdSingleEventRefresh {
			return 0, 11203, fmt.Errorf("unsupported single event cmd: %d", payload.GetCmd())
		}
		return handleSingleEventRefresh(template.ConfigData, client)
	default:
		return handleActivityOperationNoop(client)
	}
}

func handleActivityOperationNoop(client *connection.Client) (int, int, error) {
	// TODO: Implement activity operations for other activity types as needed.
	response := protobuf.SC_11203{
		Result:         proto.Uint32(0),
		AwardList:      []*protobuf.DROPINFO{},
		Build:          nil,
		Number:         []uint32{},
		ReturnUserList: []*protobuf.RETURN_USER_INFO{},
		CollectionList: nil,
		TaskList:       nil,
	}
	return client.SendMessage(11203, &response)
}

func handleSingleEventRefresh(configData json.RawMessage, client *connection.Client) (int, int, error) {
	ids, err := parseActivityConfigIDs(configData)
	if err != nil {
		return 0, 11203, err
	}

	dailyIDs := make([]uint32, 0, len(ids))
	for _, id := range ids {
		entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/activity_single_event.json", fmt.Sprintf("%d", id))
		if err != nil {
			return 0, 11203, err
		}
		var config singleEventConfig
		if err := json.Unmarshal(entry.Data, &config); err != nil {
			return 0, 11203, err
		}
		if config.Type == 2 {
			dailyIDs = append(dailyIDs, id)
		}
	}

	response := protobuf.SC_11203{
		Result:         proto.Uint32(0),
		AwardList:      []*protobuf.DROPINFO{},
		Build:          nil,
		Number:         dailyIDs,
		ReturnUserList: []*protobuf.RETURN_USER_INFO{},
		CollectionList: nil,
		TaskList:       nil,
	}
	return client.SendMessage(11203, &response)
}

func parseActivityConfigIDs(configData json.RawMessage) ([]uint32, error) {
	var ids []uint32
	if err := json.Unmarshal(configData, &ids); err == nil {
		return ids, nil
	}
	var rawIDs []any
	if err := json.Unmarshal(configData, &rawIDs); err != nil {
		return nil, err
	}
	ids = make([]uint32, 0, len(rawIDs))
	for _, value := range rawIDs {
		number, ok := parseJSONUint(value)
		if !ok {
			return nil, fmt.Errorf("unsupported activity config id: %v", value)
		}
		ids = append(ids, number)
	}
	return ids, nil
}

func parseJSONUint(value any) (uint32, bool) {
	if number, ok := value.(float64); ok {
		return uint32(number), true
	}
	if number, ok := value.(int); ok {
		return uint32(number), true
	}
	if number, ok := value.(uint32); ok {
		return number, true
	}
	return 0, false
}
