package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

func CommanderCommissionsFleet(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_13201
	response.EliteExpeditionCount = proto.Uint32(0)
	response.EscortExpeditionCount = proto.Uint32(0)
	return client.SendMessage(13201, &response)
}
