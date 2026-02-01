package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateStoryList(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11032
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11033, err
	}
	response := protobuf.SC_11033{Result: proto.Uint32(0)}
	for _, storyID := range payload.GetStoryIds() {
		if err := orm.AddCommanderStory(orm.GormDB, client.Commander.CommanderID, storyID); err != nil {
			response.Result = proto.Uint32(1)
			break
		}
	}
	return client.SendMessage(11033, &response)
}
