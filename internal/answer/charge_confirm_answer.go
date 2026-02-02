package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChargeConfirmAnswer(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11504
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11505, err
	}
	response := protobuf.SC_11505{
		Result:  proto.Uint32(5002),
		ShopId:  proto.Uint32(0),
		Gem:     proto.Uint32(0),
		GemFree: proto.Uint32(0),
	}
	return client.SendMessage(11505, &response)
}
