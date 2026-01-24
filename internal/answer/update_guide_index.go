package answer

import (
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

func UpdateGuideIndex(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_11016
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 11018, err
	}
	response := protobuf.SC_11018{
		Result:   proto.Uint32(0),
		DropList: []*protobuf.DROPINFO{},
	}
	updates := map[string]interface{}{}
	if payload.GetType() == 1 {
		client.Commander.NewGuideIndex = payload.GetGuideIndex()
		updates["new_guide_index"] = payload.GetGuideIndex()
	} else {
		client.Commander.GuideIndex = payload.GetGuideIndex()
		updates["guide_index"] = payload.GetGuideIndex()
	}
	if err := orm.GormDB.Model(client.Commander).Updates(updates).Error; err != nil {
		response.Result = proto.Uint32(1)
	}
	return client.SendMessage(11018, &response)
}
