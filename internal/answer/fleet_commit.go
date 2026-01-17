package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func FleetCommit(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_12102

	response := protobuf.SC_12103{
		Result: proto.Uint32(0),
	}

	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 12103, err
	}
	fleet, ok := client.Commander.FleetsMap[payload.GetId()]
	if !ok {
		// Create the fleet
		if err := orm.CreateFleet(client.Commander, payload.GetId(), "", payload.ShipList); err != nil {
			response.Result = proto.Uint32(1)
		}
	} else {
		// Update the fleet
		if err := fleet.UpdateShipList(client.Commander, payload.ShipList); err != nil {
			response.Result = proto.Uint32(1)
		}
	}
	return client.SendMessage(12103, &response)
}
