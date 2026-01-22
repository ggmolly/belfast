package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func JuustagramReadTip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11720
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11721, err
	}
	if client.Commander == nil {
		return 0, 11721, errors.New("missing commander")
	}
	if err := orm.MarkJuustagramChatGroupsRead(client.Commander.CommanderID, payload.GetChatGroupIdList()); err != nil {
		return 0, 11721, err
	}
	response := protobuf.SC_11721{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(11721, &response)
}
