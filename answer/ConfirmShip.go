package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func ConfirmShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	return client.SendMessage(12046, &protobuf.SC_12046{
		Result: proto.Uint32(0),
	})
}
