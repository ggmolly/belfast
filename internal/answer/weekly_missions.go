package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func WeeklyMissions(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_20101
	response.Info = &protobuf.WEEKLY_INFO{
		Pt:       proto.Uint32(0),
		RewardLv: proto.Uint32(0),
	}
	return client.SendMessage(20101, &response)
}
