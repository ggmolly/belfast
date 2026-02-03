package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ShipAction12020(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12020
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12021, err
	}

	response := protobuf.SC_12021{Result: proto.Uint32(1)}
	if _, ok := client.Commander.OwnedShipsMap[data.GetShipId()]; ok {
		response.Result = proto.Uint32(0)
	}

	return client.SendMessage(12021, &response)
}
