package answer

import (
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ZanShipEvaluation(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17105
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17106, err
	}

	response := protobuf.SC_17106{Result: proto.Uint32(0)}
	if client.Commander == nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17106, &response)
	}

	shipGroupID := payload.GetShipGroupId()
	discussID := payload.GetDiscussId()
	goodOrBad := payload.GetGoodOrBad()
	if shipGroupID == 0 || discussID == 0 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17106, &response)
	}
	if goodOrBad != 0 && goodOrBad != 1 {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17106, &response)
	}

	state := getShipDiscussState(shipGroupID, time.Now())
	commanderID := client.Commander.CommanderID

	state.mu.Lock()
	defer state.mu.Unlock()

	var entry *protobuf.DISCUSS_INFO
	for _, e := range state.discussList {
		if e.GetId() == discussID {
			entry = e
			break
		}
	}
	if entry == nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17106, &response)
	}

	if state.reviewedDiscussByCommander != nil {
		if votedByCommander, ok := state.reviewedDiscussByCommander[commanderID]; ok {
			if _, ok := votedByCommander[discussID]; ok {
				response.Result = proto.Uint32(7)
				return client.SendMessage(17106, &response)
			}
		}
	}

	if goodOrBad == 0 {
		entry.GoodCount = proto.Uint32(entry.GetGoodCount() + 1)
	} else {
		entry.BadCount = proto.Uint32(entry.GetBadCount() + 1)
	}

	if state.reviewedDiscussByCommander == nil {
		state.reviewedDiscussByCommander = map[uint32]map[uint32]struct{}{}
	}
	if state.reviewedDiscussByCommander[commanderID] == nil {
		state.reviewedDiscussByCommander[commanderID] = map[uint32]struct{}{}
	}
	state.reviewedDiscussByCommander[commanderID][discussID] = struct{}{}

	return client.SendMessage(17106, &response)
}
