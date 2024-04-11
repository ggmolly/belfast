package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC33001 protobuf.SC_33001

func UNK_33001(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(33001, &validSC33001)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC33001)
}
