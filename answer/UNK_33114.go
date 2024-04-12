package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_33114(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_33114
	response.IsWorldOpen = proto.Uint32(0)
	response.Progress = proto.Uint32(0)
	return client.SendMessage(33114, &response)
}
