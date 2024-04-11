package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC11210 protobuf.SC_11210

func PermanentActivites(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_11210
	response.PermanentNow = proto.Uint32(0)
	return client.SendMessage(11210, &response)
}

func init() {
	data := []byte{0x08, 0xfb, 0x2e, 0x08, 0xf0, 0x2e, 0x10, 0x00}
	proto.Unmarshal(data, &validSC11210)
}
