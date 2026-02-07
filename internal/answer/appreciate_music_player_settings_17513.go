package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateAppreciationMusicPlayerSettings(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17513
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17514, err
	}

	response := protobuf.SC_17514{Result: proto.Uint32(0)}
	state, err := orm.GetOrCreateCommanderAppreciationState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17514, &response)
	}
	state.MusicNo = payload.GetMusicNo()
	state.MusicMode = payload.GetMusicMode()
	if err := orm.SaveCommanderAppreciationState(orm.GormDB, state); err != nil {
		response.Result = proto.Uint32(1)
		return client.SendMessage(17514, &response)
	}

	return client.SendMessage(17514, &response)
}
