package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func InstagramMessageRange(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11705
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11706, err
	}

	response := protobuf.SC_11706{
		InsMessageList: []*protobuf.INS_MESSAGE{},
	}
	return client.SendMessage(11706, &response)
}
