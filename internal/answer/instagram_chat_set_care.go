package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func InstagramChatSetCare(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11716
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11717, err
	}
	if client.Commander == nil {
		return 0, 11717, errors.New("missing commander")
	}
	response := protobuf.SC_11717{Result: proto.Uint32(0)}
	value := payload.GetValue()
	if err := orm.UpdateJuustagramGroup(client.Commander.CommanderID, payload.GetGroupId(), nil, &value, nil); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11717, &response)
}
