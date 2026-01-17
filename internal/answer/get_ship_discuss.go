package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetShipDiscuss(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17101
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17102, err
	}
	var response protobuf.SC_17102

	response.ShipDiscuss = &protobuf.SHIP_DISCUSS_INFO{
		ShipGroupId:       payload.ShipGroupId,
		DiscussCount:      proto.Uint32(0),
		HeartCount:        proto.Uint32(0),
		DailyDiscussCount: proto.Uint32(0),
	}

	return client.SendMessage(17102, &response)
}
