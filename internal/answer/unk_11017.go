package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UNK_11017(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_11018{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	return client.SendMessage(11018, &response)
}
