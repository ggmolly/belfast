package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ToggleAppreciationMusicLike(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17507
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17508, err
	}

	action := payload.GetAction()
	switch action {
	case 0:
		if err := orm.SetCommanderAppreciationMusicFavor(orm.GormDB, client.Commander.CommanderID, payload.GetId(), true); err != nil {
			return 0, 17508, err
		}
	case 1:
		if err := orm.SetCommanderAppreciationMusicFavor(orm.GormDB, client.Commander.CommanderID, payload.GetId(), false); err != nil {
			return 0, 17508, err
		}
	}

	response := protobuf.SC_17508{Result: proto.Uint32(0)}
	return client.SendMessage(17508, &response)
}
