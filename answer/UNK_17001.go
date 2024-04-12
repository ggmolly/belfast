package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_17001(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_17001{
		DailyDiscuss: proto.Uint32(0),
	}
	return client.SendMessage(17001, &response)
}
