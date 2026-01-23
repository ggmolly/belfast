package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func WorldBaseInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_33114
	response.IsWorldOpen = proto.Uint32(0)
	response.Progress = proto.Uint32(0)
	return client.SendMessage(33114, &response)
}
