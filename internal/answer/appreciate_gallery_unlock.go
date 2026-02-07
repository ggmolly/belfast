package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UnlockAppreciateGallery(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17501
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17502, err
	}

	if err := orm.SetCommanderAppreciationGalleryUnlock(orm.GormDB, client.Commander.CommanderID, payload.GetId()); err != nil {
		return 0, 17502, err
	}

	response := protobuf.SC_17502{Result: proto.Uint32(0)}
	return client.SendMessage(17502, &response)
}
