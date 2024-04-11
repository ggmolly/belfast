package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC34002 protobuf.SC_34002

func GetMetaShipsPointsResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(34002, &validSC34002)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC34002)
}
