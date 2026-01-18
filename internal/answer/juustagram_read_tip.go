package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func JuustagramReadTip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11720
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11721, err
	}
	// TODO: Persist Juustagram read-tip state once chat data is stored.
	response := protobuf.SC_11721{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(11721, &response)
}
