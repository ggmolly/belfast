package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ChangeSelectedSkin(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12202
	if err := proto.Unmarshal(*buffer, &data); err != nil {
		return 0, 12202, err
	}
	response := protobuf.SC_12203{
		Result: proto.Uint32(0),
	}

	var ship *orm.OwnedShip
	var ok bool

	// Check if the ship is in the dock
	if ship, ok = client.Commander.OwnedShipsMap[data.GetShipId()]; !ok {
		response.Result = proto.Uint32(1)
		return client.SendMessage(12203, &response)
	}

	// Check if the skin is valid, only if it's not the default skin
	if data.GetSkinId() != 0 {
		// Check if the skin is owned
		if _, ok := client.Commander.OwnedSkinsMap[data.GetSkinId()]; !ok {
			response.Result = proto.Uint32(2)
			return client.SendMessage(12203, &response)
		}
	}

	// XXX: We voluntarily ignore whether the skin matches the ship or not, we don't care
	ship.SkinID = data.GetSkinId()
	if err := ship.Update(); err != nil {
		response.Result = proto.Uint32(3)
	}

	return client.SendMessage(12203, &response)
}
