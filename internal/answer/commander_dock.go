package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

func CommanderDock(buffer *[]byte, client *connection.Client) (int, int, error) {
	response := protobuf.SC_12010{}
	// Send ships 100:
	var shipList []*protobuf.SHIPINFO
	if len(client.Commander.Ships) > 100 {
		shipSlice := client.Commander.Ships[100:]
		shipIDs := make([]uint32, len(shipSlice))
		for i, ship := range shipSlice {
			shipIDs[i] = ship.ID
		}
		flags, err := orm.ListRandomFlagShipPhantoms(client.Commander.CommanderID, shipIDs)
		if err != nil {
			return 0, 12010, err
		}
		shadows, err := orm.ListOwnedShipShadowSkins(client.Commander.CommanderID, shipIDs)
		if err != nil {
			return 0, 12010, err
		}
		shipList = orm.ToProtoOwnedShipList(shipSlice, flags, shadows)
	}
	response.ShipList = shipList
	return client.SendMessage(12010, &response)
}
