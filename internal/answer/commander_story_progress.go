package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderStoryProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_13001{
		DailyRepairCount: proto.Uint32(3),
	}
	return client.SendMessage(13001, &response)
}
