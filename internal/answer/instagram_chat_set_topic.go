package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func InstagramChatSetTopic(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11718
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11719, err
	}
	if client.Commander == nil {
		return 0, 11719, errors.New("missing commander")
	}
	response := protobuf.SC_11719{Result: proto.Uint32(0)}
	if err := orm.SetJuustagramCurrentChatGroup(client.Commander.CommanderID, payload.GetChatGroupId()); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11719, &response)
}
