package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
)

var validSC12001 protobuf.SC_12001

func PlayerDock(buffer *[]byte, client *connection.Client) (int, int, error) {
	// Send first 100 ships
	maxSlice := 101
	if len(client.Commander.Ships) < maxSlice {
		maxSlice = len(client.Commander.Ships)
	}
	shipSlice := client.Commander.Ships[:maxSlice]
	shipIDs := make([]uint32, len(shipSlice))
	for i, ship := range shipSlice {
		shipIDs[i] = ship.ID
	}
	flags, err := orm.ListRandomFlagShipPhantoms(client.Commander.CommanderID, shipIDs)
	if err != nil {
		return 0, 12001, err
	}
	shadows, err := orm.ListOwnedShipShadowSkins(client.Commander.CommanderID, shipIDs)
	if err != nil {
		return 0, 12001, err
	}
	shipList := orm.ToProtoOwnedShipList(shipSlice, flags, shadows)
	validSC12001.Shiplist = shipList
	return client.SendMessage(12001, &validSC12001)
}
