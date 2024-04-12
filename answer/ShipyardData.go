package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
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
