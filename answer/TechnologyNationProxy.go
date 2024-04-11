package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC64000 protobuf.SC_64000

func TechnologyNationProxy(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(64000, &validSC64000)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC64000)
}
