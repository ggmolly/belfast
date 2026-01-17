package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
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
