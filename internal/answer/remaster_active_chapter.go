package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func RemasterSetActiveChapter(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13501
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13502, err
	}

	state, err := orm.GetOrCreateRemasterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		return 0, 13502, err
	}
	state.ActiveChapterID = payload.GetActiveId()
	if err := orm.GormDB.Save(state).Error; err != nil {
		return 0, 13502, err
	}

	response := protobuf.SC_13502{Result: proto.Uint32(0)}
	return client.SendMessage(13502, &response)
}
