package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC34502 protobuf.SC_34502

func UNK_34502(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(34502, &validSC34502)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC34502)
}
