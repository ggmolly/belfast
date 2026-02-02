package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChargeFailed(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11510
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11511, err
	}
	response := protobuf.SC_11511{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(11511, &response)
}
