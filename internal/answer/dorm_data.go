package answer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func DormData(buffer *[]byte, client *connection.Client) (int, int, error) {
	commanderID := client.Commander.CommanderID
	if err := tickDormAndPush(client); err != nil {
		return 0, 19001, err
	}
	furnitures, err := orm.ListCommanderFurniture(client.Commander.CommanderID)
	if err != nil {
		return 0, 19001, err
	}
	state, err := orm.GetOrCreateCommanderDormState(commanderID)
	if err != nil {
		return 0, 19001, err
	}
	var template dormLevelTemplate
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/dorm_data_template.json", fmt.Sprintf("%d", state.Level))
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return 0, 19001, err
		}
	} else {
		if err := json.Unmarshal(entry.Data, &template); err != nil {
			return 0, 19001, err
		}
	}
	layouts, err := orm.ListCommanderDormFloorLayouts(commanderID)
	if err != nil {
		return 0, 19001, err
	}
	// List ships currently in dorm (train/rest).
	var dormShips []orm.OwnedShip
	if err := orm.GormDB.Where("owner_id = ? AND state IN (5,2)", commanderID).Find(&dormShips).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, 19001, err
	}
	response := protobuf.SC_19001{
		Lv:                   proto.Uint32(state.Level),
		Food:                 proto.Uint32(state.Food),
		FoodMaxIncrease:      proto.Uint32(template.Capacity),
		FoodMaxIncreaseCount: proto.Uint32(state.FoodMaxIncreaseCount),
		FloorNum:             proto.Uint32(minUint32(state.FloorNum, 3)),
		ExpPos:               proto.Uint32(state.ExpPos),
		NextTimestamp:        proto.Uint32(state.NextTimestamp),
		LoadExp:              proto.Uint32(state.LoadExp),
		LoadFood:             proto.Uint32(state.LoadFood),
		LoadTime:             proto.Uint32(state.LoadTime),
		Name:                 proto.String(client.Commander.DormName),
	}
	if len(dormShips) > 0 {
		response.ShipIdList = make([]uint32, 0, len(dormShips))
		for i := range dormShips {
			response.ShipIdList = append(response.ShipIdList, dormShips[i].ID)
		}
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
	if len(layouts) > 0 {
		response.FurniturePutList = make([]*protobuf.FURFLOORPUTINFO, 0, len(layouts))
		for _, layout := range layouts {
			if layout.Floor == 0 || layout.Floor > 3 {
				continue
			}
			var raw []map[string]any
			// Stored JSON is compatible with unmarshalling into generic maps.
			if err := json.Unmarshal(layout.FurniturePutList, &raw); err != nil {
				return 0, 19001, err
			}
			putList := make([]*protobuf.FURNITUREPUTINFO, 0, len(raw))
			for _, m := range raw {
				b, _ := json.Marshal(m)
				var tmp protobuf.FURNITUREPUTINFO
				if err := json.Unmarshal(b, &tmp); err != nil {
					// Fallback: ignore malformed entry
					continue
				}
				// Ensure required pointers are set for proto encoding.
				id := tmp.GetId()
				x := tmp.GetX()
				y := tmp.GetY()
				dir := tmp.GetDir()
				parent := tmp.GetParent()
				shipID := tmp.GetShipId()
				children := tmp.GetChild()
				putList = append(putList, &protobuf.FURNITUREPUTINFO{
					Id:     proto.String(id),
					X:      proto.Uint32(x),
					Y:      proto.Uint32(y),
					Dir:    proto.Uint32(dir),
					Child:  children,
					Parent: proto.Uint64(parent),
					ShipId: proto.Uint32(shipID),
				})
			}
			response.FurniturePutList = append(response.FurniturePutList, &protobuf.FURFLOORPUTINFO{
				Floor:            proto.Uint32(layout.Floor),
				FurniturePutList: putList,
			})
		}
	}
	// NOTE: SC_19010 pop events are pushed by tickDormAndPush().
	state.UpdatedAtUnixTimestamp = uint32(time.Now().Unix())
	_ = orm.GormDB.Save(state).Error
	return client.SendMessage(19001, &response)
}
