package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC63000 protobuf.SC_63000

func UNK_63000(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(63000, &validSC63000)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC63000)
}
