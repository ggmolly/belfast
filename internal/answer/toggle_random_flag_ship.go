package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ToggleRandomFlagShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12204
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12205, err
	}
	response := protobuf.SC_12205{Result: proto.Uint32(0)}
	flag := payload.GetFlag()
	if flag > 1 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12205, &response)
	}
	enabled := flag == 1
	if err := orm.UpdateCommanderRandomFlagShipEnabled(orm.GormDB, client.Commander.CommanderID, enabled); err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12205, &response)
	}
	client.Commander.RandomFlagShipEnabled = enabled
	return client.SendMessage(12205, &response)
}
