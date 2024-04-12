package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func FetchSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_11604{
		State:     proto.Uint32(0),
		FailCd:    proto.Uint32(0),
		FailCount: proto.Uint32(0),
		Notice:    proto.String("TEST_NOTICE_SC_11604"),
	}
	return client.SendMessage(11604, &response)
}
