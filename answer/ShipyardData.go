package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC63100 protobuf.SC_63100

func ShipyardData(buffer *[]byte, client *connection.Client) (int, int, error) {
	validSC63100.BlueprintList = []*protobuf.BLUPRINTINFO{}
	validSC63100.ColdTime = proto.Uint32(0)
	validSC63100.DailyCatchupStrengthen = proto.Uint32(0)
	validSC63100.DailyCatchupStrengthenUr = proto.Uint32(0)

	return client.SendMessage(63100, &validSC63100)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC63100)
}
