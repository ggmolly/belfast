package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ShipyardData(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_63100{
		BlueprintList:            []*protobuf.BLUPRINTINFO{},
		ColdTime:                 proto.Uint32(0),
		DailyCatchupStrengthen:   proto.Uint32(0),
		DailyCatchupStrengthenUr: proto.Uint32(0),
	}
	return client.SendMessage(63100, &response)
}
