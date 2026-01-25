package answer

import (
	"errors"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func InstagramChatReply(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11712
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11713, err
	}
	if client.Commander == nil {
		return 0, 11713, errors.New("missing commander")
	}
	response := protobuf.SC_11713{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	now := uint32(time.Now().Unix())
	updated, err := orm.AddJuustagramChatReply(
		client.Commander.CommanderID,
		payload.GetChatGroupId(),
		payload.GetChatId(),
		payload.GetValue(),
		now,
	)
	if err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11713, &response)
	}
	response.OpTime = proto.Uint32(updated.OpTime)
	drops, err := buildJuustagramRedPacketDrops(client, payload.GetValue())
	if err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(11713, &response)
	}
	response.DropList = drops
	return client.SendMessage(11713, &response)
}

func buildJuustagramRedPacketDrops(client *connection.Client, redPacketID uint32) ([]*protobuf.DROPINFO, error) {
	config, err := getJuustagramRedPacketConfig(redPacketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*protobuf.DROPINFO{}, nil
		}
		return nil, err
	}
	if len(config.Content) != 3 {
		return nil, fmt.Errorf("invalid red packet content")
	}
	dropType := config.Content[0]
	dropID := config.Content[1]
	dropNumber := config.Content[2]
	if err := applyJuustagramDrop(client, dropType, dropID, dropNumber); err != nil {
		return nil, err
	}
	return []*protobuf.DROPINFO{
		{
			Type:   proto.Uint32(dropType),
			Id:     proto.Uint32(dropID),
			Number: proto.Uint32(dropNumber),
		},
	}, nil
}

func applyJuustagramDrop(client *connection.Client, dropType uint32, dropID uint32, dropNumber uint32) error {
	switch dropType {
	case 1:
		return client.Commander.AddResource(dropID, dropNumber)
	case 2:
		return client.Commander.AddItem(dropID, dropNumber)
	case 6:
		return client.Commander.GiveSkin(dropID)
	default:
		return fmt.Errorf("unsupported drop type")
	}
}
