package answer

import (
	"github.com/ggmolly/belfast/connection"

	"github.com/ggmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func LastOnlineInfo(buffer *[]byte, client *connection.Client) (int, int, error) {
	// dont trigger the "missed you commander" activity
	response := protobuf.SC_11752{
		Active:          proto.Uint32(1),
		ReturnLv:        nil,
		ReturnTime:      nil,
		ShipNumber:      nil,
		LastOfflineTime: nil,
		Pt:              nil,
		PtStage:         nil,
		SignCnt:         nil,
		SignLastTime:    nil,
	}
	return client.SendMessage(11752, &response)
}
