package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC24021 protobuf.SC_24021

func UNK_24021(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(24021, &validSC24021)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC24021)
}
