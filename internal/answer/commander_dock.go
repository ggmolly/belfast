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
		shipList = orm.ToProtoOwnedShipList(client.Commander.Ships[100:])
	}
	response.ShipList = shipList
	return client.SendMessage(12010, &response)
}
