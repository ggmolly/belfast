package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FleetRename(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12104
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12104, err
	}
	response := protobuf.SC_12105{
		Result: proto.Uint32(0),
	}

	// Check if the commander has this fleet, if the fleet exists, rename it
	fleet, ok := client.Commander.FleetsMap[payload.GetId()]
	if !ok {
		response.Result = proto.Uint32(1)
	} else {
		if err := fleet.RenameFleet(payload.GetName()); err != nil {
			response.Result = proto.Uint32(2)
		}
	}
	return client.SendMessage(12105, &response)
}
