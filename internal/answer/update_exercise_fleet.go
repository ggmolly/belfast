package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateExerciseFleet(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_18008
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 18009, err
	}

	response := protobuf.SC_18009{Result: proto.Uint32(0)}
	vanguardIDs := payload.GetVanguardShipIdList()
	mainIDs := payload.GetMainShipIdList()
	if len(vanguardIDs) == 0 || len(mainIDs) == 0 || len(vanguardIDs) > 3 || len(mainIDs) > 3 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(18009, &response)
	}

	for _, shipID := range vanguardIDs {
		if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
			response.Result = proto.Uint32(1)
			return client.SendMessage(18009, &response)
		}
	}
	for _, shipID := range mainIDs {
		if _, ok := client.Commander.OwnedShipsMap[shipID]; !ok {
			response.Result = proto.Uint32(1)
			return client.SendMessage(18009, &response)
		}
	}

	if err := orm.UpsertExerciseFleet(orm.GormDB, client.Commander.CommanderID, vanguardIDs, mainIDs); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(18009, &response)
}
