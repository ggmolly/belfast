package answer

import (
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func Dorm3dInstagramOp(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_28026
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 28027, err
	}
	if client.Commander == nil {
		return 0, 28027, errors.New("missing commander")
	}
	now := uint32(time.Now().Unix())
	if err := orm.UpdateDorm3dInstagramFlags(
		client.Commander.CommanderID,
		payload.GetShipId(),
		payload.GetIdList(),
		payload.GetType(),
		now,
	); err != nil {
		return 0, 28027, err
	}
	response := protobuf.SC_28027{Result: proto.Uint32(0)}
	return client.SendMessage(28027, &response)
}

func Dorm3dInstagramDiscuss(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_28028
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 28029, err
	}
	if client.Commander == nil {
		return 0, 28029, errors.New("missing commander")
	}
	now := uint32(time.Now().Unix())
	if err := orm.AddDorm3dInstagramReply(
		client.Commander.CommanderID,
		payload.GetShipId(),
		payload.GetId(),
		payload.GetChatId(),
		payload.GetValue(),
		now,
	); err != nil {
		return 0, 28029, err
	}
	response := protobuf.SC_28029{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	return client.SendMessage(28029, &response)
}
