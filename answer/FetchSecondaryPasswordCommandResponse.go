package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC11604 protobuf.SC_11604

func FetchSecondaryPasswordCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(11604, &validSC11604)
}

func init() {
	data := []byte{0x08, 0x00, 0x18, 0x00, 0x20, 0x00, 0x2a, 0x00}
	proto.Unmarshal(data, &validSC11604)
}
