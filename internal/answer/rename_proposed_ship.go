package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RenameProposedShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12034
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12035, err
	}
	response := protobuf.SC_12035{
		Result: proto.Uint32(1),
	}
	// Check if the commander has this ship, and if he's married with it, and if the rename cooldown expired (30d)
	ship, ok := client.Commander.OwnedShipsMap[data.GetShipId()]
	if !ok {
		return client.SendMessage(12035, &response)
	}

	err := ship.RenameShip(data.GetName())
	if err == orm.ErrRenameInCooldown {
		response.Result = proto.Uint32(4)
		return client.SendMessage(12035, &response)
	} else if err != nil {
		return client.SendMessage(12035, &response)
	}
	response.Result = proto.Uint32(0)
	return client.SendMessage(12035, &response)
}
