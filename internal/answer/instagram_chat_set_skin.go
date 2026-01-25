package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func InstagramChatSetSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11714
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11715, err
	}
	if client.Commander == nil {
		return 0, 11715, errors.New("missing commander")
	}
	response := protobuf.SC_11715{Result: proto.Uint32(0)}
	skinID := payload.GetSkinId()
	if err := orm.UpdateJuustagramGroup(client.Commander.CommanderID, payload.GetGroupId(), &skinID, nil, nil); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11715, &response)
}
