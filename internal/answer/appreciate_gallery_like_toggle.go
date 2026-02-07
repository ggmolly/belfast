package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func ToggleAppreciationGalleryLike(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17505
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17506, err
	}

	action := payload.GetAction()
	switch action {
	case 0:
		if err := orm.SetCommanderAppreciationGalleryFavor(orm.GormDB, client.Commander.CommanderID, payload.GetId(), true); err != nil {
			return 0, 17506, err
		}
	case 1:
		if err := orm.SetCommanderAppreciationGalleryFavor(orm.GormDB, client.Commander.CommanderID, payload.GetId(), false); err != nil {
			return 0, 17506, err
		}
	}

	response := protobuf.SC_17506{Result: proto.Uint32(0)}
	return client.SendMessage(17506, &response)
}
