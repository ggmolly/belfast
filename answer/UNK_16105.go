package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC16105 protobuf.SC_16105

func UNK_16105(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(16105, &validSC16105)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC16105)
}
