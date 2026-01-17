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
	var shipList []*protobuf.SHIPINFO
	shipList = orm.ToProtoOwnedShipList(client.Commander.Ships[:maxSlice])
	validSC12001.Shiplist = shipList
	return client.SendMessage(12001, &validSC12001)
}
