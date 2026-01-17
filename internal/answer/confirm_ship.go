package answer

import (
	"github.com/ggmolly/belfast/internal/connection"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ConfirmShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(12046, &protobuf.SC_12046{
		Result: proto.Uint32(0),
	})
}
