package answer

import (
	"errors"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	chapterOpRetreat    = 0
	chapterOpMove       = 1
	chapterOpEnemyRound = 8
	chapterOpRequest    = 49
)

func ChapterOp(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_13103
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 13104, err
	}
	state, err := orm.GetChapterState(orm.GormDB, client.Commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response := protobuf.SC_13104{Result: proto.Uint32(1)}
			return client.SendMessage(13104, &response)
		}
		return 0, 13104, err
	}
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(state.State, &current); err != nil {
		return 0, 13104, err
	}
	switch payload.GetAct() {
	case chapterOpMove:
		group := findChapterGroup(&current, payload.GetGroupId())
		if group == nil {
			response := protobuf.SC_13104{Result: proto.Uint32(1)}
			return client.SendMessage(13104, &response)
		}
		template, err := loadChapterTemplate(current.GetId(), current.GetLoopFlag())
		if err != nil {
			return 0, 13104, err
		}
		if template == nil {
			response := protobuf.SC_13104{Result: proto.Uint32(1)}
			return client.SendMessage(13104, &response)
		}
		grids, err := parseChapterGrids(template.Grids)
		if err != nil {
			return 0, 13104, err
		}
		start := chapterPos{Row: group.GetPos().GetRow(), Column: group.GetPos().GetColumn()}
		end := chapterPos{Row: payload.GetActArg_1(), Column: payload.GetActArg_2()}
		path := findMovePath(grids, start, end)
		if len(path) == 0 {
			response := protobuf.SC_13104{Result: proto.Uint32(1)}
			return client.SendMessage(13104, &response)
		}
		movePath := buildMovePath(path)
		stepDelta := uint32(len(path) - 1)
		group.Pos = buildPos(end)
		group.StepCount = proto.Uint32(group.GetStepCount() + stepDelta)
		current.MoveStepCount = proto.Uint32(current.GetMoveStepCount() + stepDelta)
		stateBytes, err := proto.Marshal(&current)
		if err != nil {
			return 0, 13104, err
		}
		state.State = stateBytes
		state.ChapterID = current.GetId()
		if err := orm.UpsertChapterState(orm.GormDB, state); err != nil {
			return 0, 13104, err
		}
		response := protobuf.SC_13104{
			Result:   proto.Uint32(0),
			MovePath: movePath,
		}
		return client.SendMessage(13104, &response)
	case chapterOpRequest:
		response := protobuf.SC_13104{
			Result:       proto.Uint32(0),
			MapUpdate:    current.GetCellList(),
			ShipUpdate:   collectChapterShips(&current),
			AiList:       current.GetAiList(),
			BuffList:     current.GetBuffList(),
			CellFlagList: current.GetCellFlagList(),
		}
		return client.SendMessage(13104, &response)
	case chapterOpEnemyRound:
		current.Round = proto.Uint32(current.GetRound() + 1)
		stateBytes, err := proto.Marshal(&current)
		if err != nil {
			return 0, 13104, err
		}
		state.State = stateBytes
		state.ChapterID = current.GetId()
		if err := orm.UpsertChapterState(orm.GormDB, state); err != nil {
			return 0, 13104, err
		}
		response := protobuf.SC_13104{
			Result:       proto.Uint32(0),
			MapUpdate:    current.GetCellList(),
			ShipUpdate:   collectChapterShips(&current),
			AiList:       current.GetAiList(),
			BuffList:     current.GetBuffList(),
			CellFlagList: current.GetCellFlagList(),
		}
		return client.SendMessage(13104, &response)
	case chapterOpRetreat:
		if err := orm.DeleteChapterState(orm.GormDB, client.Commander.CommanderID); err != nil {
			return 0, 13104, err
		}
		response := protobuf.SC_13104{Result: proto.Uint32(0)}
		return client.SendMessage(13104, &response)
	default:
		response := protobuf.SC_13104{Result: proto.Uint32(1)}
		return client.SendMessage(13104, &response)
	}
}

func findChapterGroup(current *protobuf.CURRENTCHAPTERINFO, groupID uint32) *protobuf.GROUPINCHAPTER_P13 {
	for _, group := range current.GetMainGroupList() {
		if group.GetId() == groupID {
			return group
		}
	}
	for _, group := range current.GetSubmarineGroupList() {
		if group.GetId() == groupID {
			return group
		}
	}
	for _, group := range current.GetSupportGroupList() {
		if group.GetId() == groupID {
			return group
		}
	}
	return nil
}

func collectChapterShips(current *protobuf.CURRENTCHAPTERINFO) []*protobuf.SHIPINCHAPTER_P13 {
	ships := []*protobuf.SHIPINCHAPTER_P13{}
	for _, group := range current.GetMainGroupList() {
		ships = append(ships, group.GetShipList()...)
	}
	for _, group := range current.GetSubmarineGroupList() {
		ships = append(ships, group.GetShipList()...)
	}
	for _, group := range current.GetSupportGroupList() {
		ships = append(ships, group.GetShipList()...)
	}
	return ships
}

func buildMovePath(path []chapterPos) []*protobuf.CHAPTERCELLPOS_P13 {
	movePath := make([]*protobuf.CHAPTERCELLPOS_P13, 0, len(path))
	for _, pos := range path {
		movePath = append(movePath, buildPos(pos))
	}
	return movePath
}
