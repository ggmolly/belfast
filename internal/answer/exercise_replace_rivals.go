package answer

import (
	"fmt"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ExerciseReplaceRivals(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18003
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 18004, err
	}
	if payload.Type == nil {
		return 0, 18004, fmt.Errorf("CS_18003 missing required field: type")
	}

	const rivalCount = 5
	targets := make([]*protobuf.TARGETINFO, 0, rivalCount)
	for i := 0; i < rivalCount; i++ {
		id := uint32(90000000 + i + 1)
		level := uint32(1)
		name := fmt.Sprintf("Rival #%d", i+1)
		score := uint32(0)
		rank := uint32(i + 1)
		targets = append(targets, &protobuf.TARGETINFO{
			Id:    proto.Uint32(id),
			Level: proto.Uint32(level),
			Name:  proto.String(name),
			Score: proto.Uint32(score),
			Rank:  proto.Uint32(rank),
		})
	}

	response := protobuf.SC_18004{
		Result:     proto.Uint32(0),
		TargetList: targets,
	}
	return client.SendMessage(18004, &response)
}
