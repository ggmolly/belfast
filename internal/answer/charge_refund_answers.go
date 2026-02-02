package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RefundChargeCommandAnswer(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_11514{
		Result:    proto.Uint32(5002),
		PayId:     proto.String(""),
		Url:       proto.String(""),
		OrderSign: proto.String(""),
	}
	return client.SendMessage(11514, &response)
}

func ChargeConfirmCommandAnswer(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_11505{
		Result:  proto.Uint32(5002),
		ShopId:  proto.Uint32(0),
		Gem:     proto.Uint32(0),
		GemFree: proto.Uint32(0),
	}
	return client.SendMessage(11505, &response)
}

func ChargeFailedCommandAnswer(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_11511{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(11511, &response)
}
