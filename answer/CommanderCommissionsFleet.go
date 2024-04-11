package answer

import (
	"github.com/bettercallmolly/belfast/connection"

	"github.com/bettercallmolly/belfast/protobuf"
	"google.golang.org/protobuf/proto"
)

var validSC13201 protobuf.SC_13201

func CommanderCommissionsFleet(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_13201
	response.EliteExpeditionCount = proto.Uint32(0)
	response.EscortExpeditionCount = proto.Uint32(0)
	return client.SendMessage(13201, &response)
}

func init() {
	data := []byte{}
	panic("replayed packet: replace this with the actual data")
	proto.Unmarshal(data, &validSC13201)
}
