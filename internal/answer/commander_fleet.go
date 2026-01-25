package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"google.golang.org/protobuf/proto"

	"github.com/ggmolly/belfast/internal/protobuf"
)

func CommanderFleet(buffer *[]byte, client *connection.Client) (int, int, error) {
	var response protobuf.SC_12101

	for _, fleet := range client.Commander.Fleets {
		shipList := make([]uint32, len(fleet.ShipList))
		for i, ship := range fleet.ShipList {
			shipList[i] = uint32(ship)
		}
		response.GroupList = append(response.GroupList, &protobuf.GROUPINFO_P12{
			Id:         proto.Uint32(fleet.GameID),
			Name:       proto.String(fleet.Name),
			ShipList:   shipList,
			Commanders: []*protobuf.COMMANDERSINFO{},
		})
	}

	return client.SendMessage(12101, &response)
}
