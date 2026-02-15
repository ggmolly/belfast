package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	taskPacketResultSuccess uint32 = 0
	taskPacketResultFailure uint32 = 1
)

func TaskProgressEvent(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20016
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20017, err
	}

	result := taskPacketResultFailure
	if payload.EventType != nil && payload.EventTarget != nil && payload.EventCount != nil && payload.GetEventCount() > 0 {
		result = taskPacketResultSuccess
	}

	response := protobuf.SC_20017{Result: proto.Uint32(result)}
	return client.SendMessage(20017, &response)
}

func TaskProgressBatchUpdate(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20009
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20010, err
	}

	result := taskPacketResultFailure
	if len(payload.Progressinfo) > 0 {
		valid := true
		for _, update := range payload.Progressinfo {
			if update == nil || update.Id == nil || update.Progress == nil || update.GetId() == 0 || update.GetProgress() == 0 {
				valid = false
				break
			}
		}
		if valid {
			result = taskPacketResultSuccess
		}
	}

	response := protobuf.SC_20010{Result: proto.Uint32(result)}
	return client.SendMessage(20010, &response)
}

func TaskSubmit(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20006, err
	}

	result := taskPacketResultFailure
	if payload.Id != nil && payload.GetId() > 0 {
		result = taskPacketResultSuccess
	}

	response := protobuf.SC_20006{
		Result:    proto.Uint32(result),
		AwardList: []*protobuf.DROPINFO{},
	}
	return client.SendMessage(20006, &response)
}

func TaskSubmitBatch(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20011
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20012, err
	}

	ids := make([]uint32, 0, len(payload.IdList))
	for _, id := range payload.IdList {
		if id == 0 {
			continue
		}
		ids = append(ids, id)
	}

	response := protobuf.SC_20012{
		IdList:    ids,
		AwardList: []*protobuf.DROPINFO{},
	}
	return client.SendMessage(20012, &response)
}

func TaskQuickFinish(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20013
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20014, err
	}

	result := taskPacketResultFailure
	if payload.Id != nil && payload.GetId() > 0 {
		result = taskPacketResultSuccess
	}

	response := protobuf.SC_20014{
		Result:    proto.Uint32(result),
		AwardList: []*protobuf.DROPINFO{},
	}
	return client.SendMessage(20014, &response)
}
