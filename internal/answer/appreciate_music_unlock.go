package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UnlockAppreciateMusic(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17503
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17504, err
	}

	response := protobuf.SC_17504{Result: proto.Uint32(0)}
	return client.SendMessage(17504, &response)
}
