package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetMyAssaultFleetCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_61010{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(61010, &response)
}
