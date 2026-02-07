package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const maxMedalDisplayEntries = 5

func ChangeMedalDisplay(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17401
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17402, err
	}
	response := protobuf.SC_17402{Result: proto.Uint32(0)}
	if payload.GetFixedConst() != 1 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17402, &response)
	}
	medalIDs := payload.GetMedalId()
	if len(medalIDs) > maxMedalDisplayEntries {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17402, &response)
	}
	if !validateMedalDisplayList(medalIDs) {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17402, &response)
	}
	if err := orm.SetCommanderMedalDisplay(orm.GormDB, client.Commander.CommanderID, medalIDs); err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17402, &response)
	}
	return client.SendMessage(17402, &response)
}

func validateMedalDisplayList(medalIDs []uint32) bool {
	if len(medalIDs) == 0 {
		return true
	}
	seen := make(map[uint32]struct{}, len(medalIDs))
	for _, medalID := range medalIDs {
		if medalID == 0 {
			return false
		}
		if _, ok := seen[medalID]; ok {
			return false
		}
		seen[medalID] = struct{}{}
	}
	return true
}
