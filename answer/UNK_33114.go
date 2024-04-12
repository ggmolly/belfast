package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_33114(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_33114
	response.IsWorldOpen = proto.Uint32(0)
	response.Progress = proto.Uint32(0)
	return client.SendMessage(33114, &response)
}
