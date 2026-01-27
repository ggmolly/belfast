package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChangeRandomFlagShipMode(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12206
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12207, err
	}
	response := protobuf.SC_12207{Result: proto.Uint32(0)}
	mode := payload.GetFlag()
	if mode < 1 || mode > 3 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12207, &response)
	}
	if err := orm.UpdateCommanderRandomShipMode(orm.GormDB, client.Commander.CommanderID, mode); err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12207, &response)
	}
	client.Commander.RandomShipMode = mode
	return client.SendMessage(12207, &response)
}
