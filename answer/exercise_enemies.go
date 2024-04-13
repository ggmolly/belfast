package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func ExerciseEnemies(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_18002{
		Score:               proto.Uint32(0),
		Rank:                proto.Uint32(0),
		FightCount:          proto.Uint32(0),
		FightCountResetTime: proto.Uint32(0),
		FlashTargetCount:    proto.Uint32(0),
	}
	return client.SendMessage(18002, &response)
}
