package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC60101 protobuf.SC_60101

func CommanderGuildChat(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(60101, &validSC60101)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC60101)
}
