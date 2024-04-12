package answer

import (
	"github.com/bettercallmolly/belfast/connection"
	"github.com/bettercallmolly/belfast/protobuf"

	"google.golang.org/protobuf/proto"
)

func GetMetaProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_63315{
		Type: proto.Uint32(1),
	}
	return client.SendMessage(63315, &response)
}
