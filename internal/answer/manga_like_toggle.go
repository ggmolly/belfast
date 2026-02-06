package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ToggleMangaLike(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17511
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17512, err
	}

	action := payload.GetAction()
	switch action {
	case 0:
		if err := orm.SetCommanderCartoonCollectMark(orm.GormDB, client.Commander.CommanderID, payload.GetId(), true); err != nil {
			return 0, 17512, err
		}
	case 1:
		if err := orm.SetCommanderCartoonCollectMark(orm.GormDB, client.Commander.CommanderID, payload.GetId(), false); err != nil {
			return 0, 17512, err
		}
	}

	response := protobuf.SC_17512{Result: proto.Uint32(0)}
	return client.SendMessage(17512, &response)
}
