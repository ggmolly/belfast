package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func MarkMangaRead(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_17509
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 17510, err
	}

	if err := orm.SetCommanderCartoonReadMark(orm.GormDB, client.Commander.CommanderID, payload.GetId()); err != nil {
		return 0, 17510, err
	}

	response := protobuf.SC_17510{Result: proto.Uint32(0)}
	return client.SendMessage(17510, &response)
}
