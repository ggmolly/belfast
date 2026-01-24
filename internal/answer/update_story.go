package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateStory(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11017
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11018, err
	}
	response := protobuf.SC_11018{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	if err := orm.AddCommanderStory(orm.GormDB, client.Commander.CommanderID, payload.GetStoryId()); err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11018, &response)
}
