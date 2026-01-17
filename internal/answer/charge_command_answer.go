package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChargeCommandAnswer(buffer *[]byte, client *connection.Client) (int, int, error) {
	// Disable Charge
	response := protobuf.SC_11502{
		Result:    proto.Uint32(5002),
		PayId:     proto.String(""),
		Url:       proto.String(""),
		OrderSign: proto.String(""),
	}

	return client.SendMessage(11502, &response)
}
