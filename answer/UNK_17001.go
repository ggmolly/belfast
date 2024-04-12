package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_17001(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_17001{
		DailyDiscuss: proto.Uint32(0),
	}
	return client.SendMessage(17001, &response)
}
