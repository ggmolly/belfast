package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func LimitChallengeInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_24021{
		Result: proto.Uint32(0),
	}
	return client.SendMessage(24021, &response)
}
