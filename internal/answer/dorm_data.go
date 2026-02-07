package answer

import (
	"encoding/json"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type dormTemplate struct {
	Capacity uint32 `json:"capacity"`
}

func DormData(buffer *[]byte, client *connection.Client) (int, int, error) {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/dorm_data_template.json")
	if err != nil {
		return 0, 19001, err
	}
	furnitures, err := orm.ListCommanderFurniture(client.Commander.CommanderID)
	if err != nil {
		return 0, 19001, err
	}
	response := protobuf.SC_19001{
		Lv:                   proto.Uint32(0),
		Food:                 proto.Uint32(0),
		FoodMaxIncrease:      proto.Uint32(0),
		FoodMaxIncreaseCount: proto.Uint32(0),
		FloorNum:             proto.Uint32(uint32(len(entries))),
		ExpPos:               proto.Uint32(0),
		NextTimestamp:        proto.Uint32(0),
		LoadExp:              proto.Uint32(0),
		LoadFood:             proto.Uint32(0),
		LoadTime:             proto.Uint32(0),
		Name:                 proto.String(""),
	}
	if len(furnitures) > 0 {
		response.FurnitureIdList = make([]*protobuf.FURNITUREINFO, 0, len(furnitures))
		for i := range furnitures {
			furniture := furnitures[i]
			response.FurnitureIdList = append(response.FurnitureIdList, &protobuf.FURNITUREINFO{
				Id:      proto.Uint32(furniture.FurnitureID),
				Count:   proto.Uint32(furniture.Count),
				GetTime: proto.Uint32(furniture.GetTime),
			})
		}
	}
	if len(entries) > 0 {
		var template dormTemplate
		if err := json.Unmarshal(entries[0].Data, &template); err != nil {
			return 0, 19001, err
		}
		response.FoodMaxIncrease = proto.Uint32(template.Capacity)
	}
	return client.SendMessage(19001, &response)
}
