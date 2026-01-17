package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func LastOnlineInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	// dont trigger the "missed you commander" activity
	response := protobuf.SC_11752{
		Active:          proto.Uint32(0),
		ReturnLv:        proto.Uint32(0),
		ReturnTime:      proto.Uint32(0),
		ShipNumber:      proto.Uint32(0),
		LastOfflineTime: proto.Uint32(0),
		Pt:              proto.Uint32(0),
		PtStage:         proto.Uint32(0),
		SignCnt:         proto.Uint32(0),
		SignLastTime:    proto.Uint32(0),
	}
	return client.SendMessage(11752, &response)
}
