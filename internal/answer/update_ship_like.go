package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateShipLike(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17107
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17108, err
	}
	response := protobuf.SC_17108{
		Result: proto.Uint32(boolToUint32(client.Commander.Like(*payload.ShipGroupId) != nil)),
	}
	return client.SendMessage(17108, &response)
}
