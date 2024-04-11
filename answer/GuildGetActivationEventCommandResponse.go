package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC61006 protobuf.SC_61006

func GuildGetActivationEventCommandResponse(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(61006, &validSC61006)
}

func init() {
	data := []byte{0x08, 0x14}
	proto.Unmarshal(data, &validSC61006)
}
