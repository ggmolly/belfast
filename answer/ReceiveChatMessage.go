package answer

import (
	"fmt"

	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/orm"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func ReceiveChatMessage(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_50102
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 50101, fmt.Errorf("invalid CS_50102 packet: %s", err.Error())
	}
	msg, err := orm.SendMessage(client.Commander.RoomID, data.GetContent(), client.Commander)
	if err != nil {
		return 0, 50101, fmt.Errorf("unable to save message: %s", err.Error())
	}
	client.Server.SendMessage(client, *msg)
	return 0, 50101, nil
}
