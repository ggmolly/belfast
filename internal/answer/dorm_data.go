package answer

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

type dormShipSnapshot struct {
	ID     uint32
	TID    uint32
	State  uint32
	SkinID uint32
}

type dormSnapshot struct {
	State      *orm.CommanderDormState
	Template   dormLevelTemplate
	DormName   string
	Furnitures []orm.CommanderFurniture
	Layouts    []orm.CommanderDormFloorLayout
	Ships      []dormShipSnapshot
}

func DormData(buffer *[]byte, client *connection.Client) (int, int, error) {
	commanderID := client.Commander.CommanderID
	if err := tickDormAndPush(client); err != nil {
		return 0, 19001, err
	}
	snapshot, err := loadDormSnapshot(commanderID, client.Commander.DormName)
	if err != nil {
		return 0, 19001, err
	}
	response, err := buildDormDataResponse(snapshot)
	if err != nil {
		return 0, 19001, err
	}
	// NOTE: SC_19010 pop events are pushed by tickDormAndPush().
	snapshot.State.UpdatedAtUnixTimestamp = uint32(time.Now().Unix())
	_ = orm.SaveCommanderDormState(snapshot.State)
	return client.SendMessage(19001, &response)
}

func VisitBackyard19101(buffer *[]byte, client *connection.Client) (int, int, error) {
	var request protobuf.CS_19101
	if err := proto.Unmarshal(*buffer, &request); err != nil {
		return 0, 19102, err
	}

	targetCommanderID := request.GetUserId()
	targetCommander, err := orm.GetCommanderCoreByID(targetCommanderID)
	if err != nil {
		return sendVisitBackyardUnavailable(client, targetCommanderID, "target_not_found")
	}

	snapshot, err := loadDormSnapshot(targetCommanderID, targetCommander.DormName)
	if err != nil {
		return 0, 19102, err
	}

	response, err := buildVisitBackyardResponse(snapshot, targetCommander.Name)
	if err != nil {
		return 0, 19102, err
	}

	return client.SendMessage(19102, &response)
}

func sendVisitBackyardUnavailable(client *connection.Client, targetCommanderID uint32, reason string) (int, int, error) {
	logger.WithFields(
		"BackyardVisit",
		logger.FieldValue("requester", client.Commander.CommanderID),
		logger.FieldValue("target", targetCommanderID),
		logger.FieldValue("reason", reason),
	).Warn("visit backyard unavailable")

	response := protobuf.SC_19102{
		Lv:                   proto.Uint32(0),
		Food:                 proto.Uint32(0),
		FoodMaxIncrease:      proto.Uint32(0),
		FoodMaxIncreaseCount: proto.Uint32(0),
		FloorNum:             proto.Uint32(0),
		ExpPos:               proto.Uint32(0),
		Name:                 proto.String(""),
	}

	return client.SendMessage(19102, &response)
}

func loadDormSnapshot(commanderID uint32, dormName string) (*dormSnapshot, error) {
	furnitures, err := orm.ListCommanderFurniture(commanderID)
	if err != nil {
		return nil, err
	}

	state, err := orm.GetOrCreateCommanderDormState(commanderID)
	if err != nil {
		return nil, err
	}

	template := dormLevelTemplate{}
	loadedTemplate, err := loadDormLevelTemplate(state.Level)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, err
		}
	} else {
		template = *loadedTemplate
	}

	layouts, err := orm.ListCommanderDormFloorLayouts(commanderID)
	if err != nil {
		return nil, err
	}

	ships, err := listDormShipsSnapshot(commanderID)
	if err != nil {
		return nil, err
	}

	return &dormSnapshot{
		State:      state,
		Template:   template,
		DormName:   dormName,
		Furnitures: furnitures,
		Layouts:    layouts,
		Ships:      ships,
	}, nil
}

func buildDormDataResponse(snapshot *dormSnapshot) (protobuf.SC_19001, error) {
	response := protobuf.SC_19001{
		Lv:                   proto.Uint32(snapshot.State.Level),
		Food:                 proto.Uint32(snapshot.State.Food),
		FoodMaxIncrease:      proto.Uint32(snapshot.Template.Capacity),
		FoodMaxIncreaseCount: proto.Uint32(snapshot.State.FoodMaxIncreaseCount),
		FloorNum:             proto.Uint32(minUint32(snapshot.State.FloorNum, 3)),
		ExpPos:               proto.Uint32(snapshot.State.ExpPos),
		NextTimestamp:        proto.Uint32(snapshot.State.NextTimestamp),
		LoadExp:              proto.Uint32(snapshot.State.LoadExp),
		LoadFood:             proto.Uint32(snapshot.State.LoadFood),
		LoadTime:             proto.Uint32(snapshot.State.LoadTime),
		Name:                 proto.String(snapshot.DormName),
	}

	if len(snapshot.Ships) > 0 {
		shipIDs := make([]uint32, 0, len(snapshot.Ships))
		for _, ship := range snapshot.Ships {
			shipIDs = append(shipIDs, ship.ID)
		}
		response.ShipIdList = shipIDs
	}

	response.FurnitureIdList = buildDormFurnitureInfoList(snapshot.Furnitures)

	floorPutList, err := buildDormFloorPutList(snapshot.Layouts)
	if err != nil {
		return protobuf.SC_19001{}, err
	}
	response.FurniturePutList = floorPutList

	return response, nil
}

func buildVisitBackyardResponse(snapshot *dormSnapshot, name string) (protobuf.SC_19102, error) {
	response := protobuf.SC_19102{
		Lv:                   proto.Uint32(snapshot.State.Level),
		Food:                 proto.Uint32(snapshot.State.Food),
		FoodMaxIncrease:      proto.Uint32(snapshot.Template.Capacity),
		FoodMaxIncreaseCount: proto.Uint32(snapshot.State.FoodMaxIncreaseCount),
		FloorNum:             proto.Uint32(minUint32(snapshot.State.FloorNum, 3)),
		ExpPos:               proto.Uint32(snapshot.State.ExpPos),
		Name:                 proto.String(name),
	}

	if len(snapshot.Ships) > 0 {
		response.ShipIdList = make([]*protobuf.SHIP_IN_DROM, 0, len(snapshot.Ships))
		for _, ship := range snapshot.Ships {
			response.ShipIdList = append(response.ShipIdList, &protobuf.SHIP_IN_DROM{
				Id:     proto.Uint32(ship.ID),
				Tid:    proto.Uint32(ship.TID),
				State:  proto.Uint32(ship.State),
				SkinId: proto.Uint32(ship.SkinID),
			})
		}
	}

	response.FurnitureIdList = buildDormFurnitureInfoList(snapshot.Furnitures)

	floorPutList, err := buildDormFloorPutList(snapshot.Layouts)
	if err != nil {
		return protobuf.SC_19102{}, err
	}
	response.FurniturePutList = floorPutList

	return response, nil
}

func buildDormFurnitureInfoList(furnitures []orm.CommanderFurniture) []*protobuf.FURNITUREINFO {
	if len(furnitures) == 0 {
		return nil
	}

	result := make([]*protobuf.FURNITUREINFO, 0, len(furnitures))
	for i := range furnitures {
		furniture := furnitures[i]
		result = append(result, &protobuf.FURNITUREINFO{
			Id:      proto.Uint32(furniture.FurnitureID),
			Count:   proto.Uint32(furniture.Count),
			GetTime: proto.Uint32(furniture.GetTime),
		})
	}

	return result
}

func buildDormFloorPutList(layouts []orm.CommanderDormFloorLayout) ([]*protobuf.FURFLOORPUTINFO, error) {
	if len(layouts) == 0 {
		return nil, nil
	}

	result := make([]*protobuf.FURFLOORPUTINFO, 0, len(layouts))
	for _, layout := range layouts {
		if layout.Floor == 0 || layout.Floor > 3 {
			continue
		}

		var raw []map[string]any
		if err := json.Unmarshal(layout.FurniturePutList, &raw); err != nil {
			return nil, err
		}

		putList := make([]*protobuf.FURNITUREPUTINFO, 0, len(raw))
		for _, entry := range raw {
			encodedEntry, _ := json.Marshal(entry)
			var tmp protobuf.FURNITUREPUTINFO
			if err := json.Unmarshal(encodedEntry, &tmp); err != nil {
				continue
			}

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

		result = append(result, &protobuf.FURFLOORPUTINFO{
			Floor:            proto.Uint32(layout.Floor),
			FurniturePutList: putList,
		})
	}

	return result, nil
}

func listDormShipsSnapshot(commanderID uint32) ([]dormShipSnapshot, error) {
	rows, err := db.DefaultStore.Pool.Query(context.Background(), `
SELECT id, ship_id, state, skin_id
FROM owned_ships
WHERE owner_id = $1
  AND deleted_at IS NULL
  AND state IN (5, 2)
`, int64(commanderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ships := make([]dormShipSnapshot, 0)
	for rows.Next() {
		var id int64
		var tid int64
		var state int64
		var skinID int64
		if err := rows.Scan(&id, &tid, &state, &skinID); err != nil {
			return nil, err
		}
		ships = append(ships, dormShipSnapshot{
			ID:     uint32(id),
			TID:    uint32(tid),
			State:  uint32(state),
			SkinID: uint32(skinID),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ships, nil
}
