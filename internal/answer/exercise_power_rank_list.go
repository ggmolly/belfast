package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ExercisePowerRankList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18006
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 18007, err
	}
	if payload.Type == nil {
		return 0, 18007, fmt.Errorf("CS_18006 missing required field: type")
	}

	const rankCount = 5
	ranks := make([]*protobuf.ARENARANK, 0, rankCount)

	if client.Commander != nil {
		ranks = append(ranks, &protobuf.ARENARANK{
			Id:    proto.Uint32(client.Commander.CommanderID),
			Level: proto.Uint32(uint32(client.Commander.Level)),
			Name:  proto.String(client.Commander.Name),
			Score: proto.Uint32(0),
		})
	}

	for i := len(ranks); i < rankCount; i++ {
		id := uint32(91000000 + i + 1)
		name := fmt.Sprintf("Commander #%d", i+1)
		ranks = append(ranks, &protobuf.ARENARANK{
			Id:    proto.Uint32(id),
			Level: proto.Uint32(1),
			Name:  proto.String(name),
			Score: proto.Uint32(0),
		})
	}

	response := protobuf.SC_18007{
		ArenaRankLsit: ranks,
	}
	return client.SendMessage(18007, &response)
}
