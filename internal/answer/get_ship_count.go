package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetShipCount(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11800
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11801, err
	}

	response := protobuf.SC_11801{
		ShipCount: proto.Uint32(uint32(len(client.Commander.Ships))),
	}

	return client.SendMessage(11801, &response)
}
