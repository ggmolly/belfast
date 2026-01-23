package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func WorldCheckInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_33001{
		IsWorldOpen: proto.Uint32(0),
		Camp:        proto.Uint32(0),
		CountInfo: &protobuf.COUNTINFO{
			StepCount:     proto.Uint32(0),
			TreasureCount: proto.Uint32(0),
			TaskProgress:  proto.Uint32(0),
			ActivateCount: proto.Uint32(0),
		},
	}
	return client.SendMessage(33001, &response)
}
