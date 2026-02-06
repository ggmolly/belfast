package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func GetShip(buffer *[]byte, client *connection.Client) (int, int, error) {
	var data protobuf.CS_12025
	err := proto.Unmarshal(*buffer, &data)
	if err != nil {
		return 0, 12025, err
	}

	response := protobuf.SC_12026{
		Result: proto.Uint32(0),
	}

	response.ShipList = make([]*protobuf.SHIPINFO, len(data.GetPosList()))
	var minPos uint32 = 999999
	var maxPos uint32
	for _, pos := range data.GetPosList() {
		if pos < minPos {
			minPos = pos
		}
		if pos > maxPos {
			maxPos = pos
		}
	}
	// since the game is using lua, the indexes start at 1
	minPos -= 1
	maxPos -= 1
	if maxPos == minPos {
		maxPos += 1
	}
	builds, err := client.Commander.GetBuildRange(minPos, maxPos)
	if err != nil {
		return 0, 12025, err
	}

	ships := make([]*orm.OwnedShip, len(data.GetPosList()))
	shipIDs := make([]uint32, len(data.GetPosList()))
	for i := range data.GetPosList() {
		ship, err := builds[i].Consume(builds[i].ShipID, client.Commander)
		if err != nil {
			return 0, 12025, err
		}
		ships[i] = ship
		shipIDs[i] = ship.ID
	}
	flags, err := orm.ListRandomFlagShipPhantoms(client.Commander.CommanderID, shipIDs)
	if err != nil {
		return 0, 12025, err
	}
	shadows, err := orm.ListOwnedShipShadowSkins(client.Commander.CommanderID, shipIDs)
	if err != nil {
		return 0, 12025, err
	}
	for i, ship := range ships {
		response.ShipList[i] = orm.ToProtoOwnedShip(*ship, flags[ship.ID], shadows[ship.ID])
	}

	return client.SendMessage(12026, &response)
}
