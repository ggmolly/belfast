package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RefundChargeAnswer(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11513
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11514, err
	}
	response := protobuf.SC_11514{
		Result:    proto.Uint32(5002),
		PayId:     proto.String(""),
		Url:       proto.String(""),
		OrderSign: proto.String(""),
	}
	return client.SendMessage(11514, &response)
}
