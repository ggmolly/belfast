package answer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	chapterAttachBorn         = 1
	chapterAttachBornSub      = 16
	chapterAttachBoss         = 8
	chapterAttachElite        = 4
	chapterAttachAmbush       = 5
	chapterAttachEnemy        = 6
	chapterAttachTorpedoEnemy = 7
	chapterAttachChampion     = 12
	chapterAttachBombEnemy    = 24
	chapterCellActive         = 0
	chapterCellDisabled       = 1
	chapterCellAmbush         = 2
)

type chapterGrid struct {
	Row        uint32
	Column     uint32
	Walkable   bool
	Attachment uint32
}

type chapterPos struct {
	Row    uint32
	Column uint32
}

func buildCurrentChapterInfo(template *chapterTemplate, payload *protobuf.CS_13101, operationBuffID uint32) (*protobuf.CURRENTCHAPTERINFO, uint32, error) {
	grids, err := parseChapterGrids(template.Grids)
	if err != nil {
		return nil, 0, err
	}
	mainSpawns := selectSpawnPositions(grids, chapterAttachBorn)
	subSpawns := selectSpawnPositions(grids, chapterAttachBornSub)
	cellList := buildChapterCells(grids, template)
	mainGroups, mainCount := buildGroupsFromTeams(payload.GetFleet().GetMainTeam(), mainSpawns, template.AmmoTotal)
	subGroups, subCount := buildGroupsFromTeams(payload.GetFleet().GetSubmarineTeam(), subSpawns, template.AmmoSubmarine)
	supportGroups, supportCount := buildGroupsFromTeams(payload.GetFleet().GetSupportTeam(), mainSpawns, template.AmmoTotal)
	strategies := buildChapterStrategies(template.ChapterStrategy)
	initShipCount := mainCount + subCount + supportCount
	current := &protobuf.CURRENTCHAPTERINFO{
		Id:                    proto.Uint32(payload.GetId()),
		Time:                  proto.Uint32(uint32(time.Now().Unix()) + template.Time),
		CellList:              cellList,
		MainGroupList:         mainGroups,
		AiList:                []*protobuf.CHAPTERCELLINFO_P13{},
		EscortList:            []*protobuf.CHAPTERCELLINFO_P13{},
		Round:                 proto.Uint32(0),
		IsSubmarineAutoAttack: proto.Uint32(0),
		OperationBuff:         buildOperationBuffList(operationBuffID),
		ModelActCount:         proto.Uint32(0),
		BuffList:              []uint32{},
		LoopFlag:              proto.Uint32(payload.GetLoopFlag()),
		ExtraFlagList:         []uint32{},
		CellFlagList:          []*protobuf.CELLFLAG{},
		ChapterHp:             proto.Uint32(0),
		ChapterStrategyList:   strategies,
		KillCount:             proto.Uint32(0),
		InitShipCount:         proto.Uint32(initShipCount),
		ContinuousKillCount:   proto.Uint32(0),
		BattleStatistics:      []*protobuf.STRATEGYINFO_P13{},
		FleetDuties:           payload.GetFleetDuties(),
		MoveStepCount:         proto.Uint32(0),
		SubmarineGroupList:    subGroups,
		SupportGroupList:      supportGroups,
	}
	return current, initShipCount, nil
}

func buildCurrentChapterInfoKR(template *chapterTemplate, payload *protobuf.CS_13101_KR, operationBuffID uint32) (*protobuf.CURRENTCHAPTERINFO, uint32, error) {
	grids, err := parseChapterGrids(template.Grids)
	if err != nil {
		return nil, 0, err
	}
	mainSpawns := selectSpawnPositions(grids, chapterAttachBorn)
	cellList := buildChapterCells(grids, template)
	mainGroups, mainCount := buildGroupsFromElite(payload.GetGroupIdList(), payload.GetEliteFleetList(), mainSpawns, template.AmmoTotal)
	strategies := buildChapterStrategies(template.ChapterStrategy)
	current := &protobuf.CURRENTCHAPTERINFO{
		Id:                    proto.Uint32(payload.GetId()),
		Time:                  proto.Uint32(uint32(time.Now().Unix()) + template.Time),
		CellList:              cellList,
		MainGroupList:         mainGroups,
		AiList:                []*protobuf.CHAPTERCELLINFO_P13{},
		EscortList:            []*protobuf.CHAPTERCELLINFO_P13{},
		Round:                 proto.Uint32(0),
		IsSubmarineAutoAttack: proto.Uint32(0),
		OperationBuff:         buildOperationBuffList(operationBuffID),
		ModelActCount:         proto.Uint32(0),
		BuffList:              []uint32{},
		LoopFlag:              proto.Uint32(payload.GetLoopFlag()),
		ExtraFlagList:         []uint32{},
		CellFlagList:          []*protobuf.CELLFLAG{},
		ChapterHp:             proto.Uint32(0),
		ChapterStrategyList:   strategies,
		KillCount:             proto.Uint32(0),
		InitShipCount:         proto.Uint32(mainCount),
		ContinuousKillCount:   proto.Uint32(0),
		BattleStatistics:      []*protobuf.STRATEGYINFO_P13{},
		FleetDuties:           payload.GetFleetDuties(),
		MoveStepCount:         proto.Uint32(0),
		SubmarineGroupList:    []*protobuf.GROUPINCHAPTER_P13{},
		SupportGroupList:      []*protobuf.GROUPINCHAPTER_P13{},
	}
	return current, mainCount, nil
}

func buildOperationBuffList(buffID uint32) []uint32 {
	if buffID == 0 {
		return []uint32{}
	}
	return []uint32{buffID}
}

func buildChapterStrategies(ids []uint32) []*protobuf.STRATEGYINFO_P13 {
	if len(ids) == 0 {
		return []*protobuf.STRATEGYINFO_P13{}
	}
	strategies := make([]*protobuf.STRATEGYINFO_P13, 0, len(ids))
	for _, id := range ids {
		strategies = append(strategies, &protobuf.STRATEGYINFO_P13{
			Id:    proto.Uint32(id),
			Count: proto.Uint32(0),
		})
	}
	return strategies
}

func buildGroupsFromTeams(teams []*protobuf.TEAM_INFO, spawns []chapterPos, ammo uint32) ([]*protobuf.GROUPINCHAPTER_P13, uint32) {
	groups := make([]*protobuf.GROUPINCHAPTER_P13, 0, len(teams))
	var shipCount uint32
	for index, team := range teams {
		spawn := chooseSpawn(spawns, index)
		ships := make([]*protobuf.SHIPINCHAPTER_P13, 0, len(team.GetShipList()))
		for _, shipID := range team.GetShipList() {
			ships = append(ships, &protobuf.SHIPINCHAPTER_P13{
				Id:     proto.Uint32(shipID),
				HpRant: proto.Uint32(10000),
			})
			shipCount++
		}
		commanders := buildCommanderList(team.GetCommanderMain(), team.GetCommanderSub())
		groupID := team.GetId()
		if groupID == 0 {
			groupID = uint32(index + 1)
		}
		groups = append(groups, &protobuf.GROUPINCHAPTER_P13{
			Id:               proto.Uint32(groupID),
			ShipList:         ships,
			Pos:              buildPos(spawn),
			StepCount:        proto.Uint32(0),
			BoxStrategyList:  []*protobuf.STRATEGYINFO_P13{},
			ShipStrategyList: []*protobuf.STRATEGYINFO_P13{},
			StrategyIds:      []uint32{},
			Bullet:           proto.Uint32(ammo),
			StartPos:         buildPos(spawn),
			CommanderList:    commanders,
			MoveStepDown:     proto.Uint32(0),
			KillCount:        proto.Uint32(0),
			FleetId:          proto.Uint32(groupID),
			VisionLv:         proto.Uint32(0),
		})
	}
	return groups, shipCount
}

func buildGroupsFromElite(groupIDs []uint32, elite []*protobuf.ELITEFLEETINFO, spawns []chapterPos, ammo uint32) ([]*protobuf.GROUPINCHAPTER_P13, uint32) {
	groups := make([]*protobuf.GROUPINCHAPTER_P13, 0, len(groupIDs))
	var shipCount uint32
	for index, groupID := range groupIDs {
		spawn := chooseSpawn(spawns, index)
		var eliteFleet *protobuf.ELITEFLEETINFO
		if index < len(elite) {
			eliteFleet = elite[index]
		}
		ships := []*protobuf.SHIPINCHAPTER_P13{}
		if eliteFleet != nil {
			ships = make([]*protobuf.SHIPINCHAPTER_P13, 0, len(eliteFleet.GetShipIdList()))
			for _, shipID := range eliteFleet.GetShipIdList() {
				ships = append(ships, &protobuf.SHIPINCHAPTER_P13{
					Id:     proto.Uint32(shipID),
					HpRant: proto.Uint32(10000),
				})
				shipCount++
			}
		}
		commanders := []*protobuf.COMMANDERSINFO{}
		if eliteFleet != nil {
			commanders = make([]*protobuf.COMMANDERSINFO, 0, len(eliteFleet.GetCommanders()))
			for _, commander := range eliteFleet.GetCommanders() {
				commanders = append(commanders, &protobuf.COMMANDERSINFO{
					Pos: proto.Uint32(commander.GetPos()),
					Id:  proto.Uint32(commander.GetId()),
				})
			}
		}
		if groupID == 0 {
			groupID = uint32(index + 1)
		}
		groups = append(groups, &protobuf.GROUPINCHAPTER_P13{
			Id:               proto.Uint32(groupID),
			ShipList:         ships,
			Pos:              buildPos(spawn),
			StepCount:        proto.Uint32(0),
			BoxStrategyList:  []*protobuf.STRATEGYINFO_P13{},
			ShipStrategyList: []*protobuf.STRATEGYINFO_P13{},
			StrategyIds:      []uint32{},
			Bullet:           proto.Uint32(ammo),
			StartPos:         buildPos(spawn),
			CommanderList:    commanders,
			MoveStepDown:     proto.Uint32(0),
			KillCount:        proto.Uint32(0),
			FleetId:          proto.Uint32(groupID),
			VisionLv:         proto.Uint32(0),
		})
	}
	return groups, shipCount
}

func buildCommanderList(mainID uint32, subID uint32) []*protobuf.COMMANDERSINFO {
	commanders := []*protobuf.COMMANDERSINFO{}
	if mainID != 0 {
		commanders = append(commanders, &protobuf.COMMANDERSINFO{
			Pos: proto.Uint32(1),
			Id:  proto.Uint32(mainID),
		})
	}
	if subID != 0 {
		commanders = append(commanders, &protobuf.COMMANDERSINFO{
			Pos: proto.Uint32(2),
			Id:  proto.Uint32(subID),
		})
	}
	return commanders
}

func buildChapterCells(grids []chapterGrid, template *chapterTemplate) []*protobuf.CHAPTERCELLINFO_P13 {
	if len(grids) == 0 {
		return []*protobuf.CHAPTERCELLINFO_P13{}
	}
	cells := make([]*protobuf.CHAPTERCELLINFO_P13, 0, len(grids))
	var bossID uint32
	if template != nil && len(template.BossExpeditionID) > 0 {
		bossID = template.BossExpeditionID[0]
	}
	for _, grid := range grids {
		if grid.Attachment == 0 {
			continue
		}
		cell := &protobuf.CHAPTERCELLINFO_P13{
			Pos:      buildPos(chapterPos{Row: grid.Row, Column: grid.Column}),
			ItemType: proto.Uint32(grid.Attachment),
			ItemFlag: proto.Uint32(resolveCellFlag(grid.Attachment)),
			ItemData: proto.Uint32(0),
		}
		if grid.Attachment == chapterAttachBoss && bossID != 0 {
			cell.ItemId = proto.Uint32(bossID)
		} else if template != nil {
			attachmentID := selectAttachmentID(grid.Attachment, template)
			if attachmentID != 0 {
				cell.ItemId = proto.Uint32(attachmentID)
			}
		}
		cells = append(cells, cell)
	}
	return cells
}

func resolveCellFlag(attachment uint32) uint32 {
	if attachment == chapterAttachAmbush {
		return chapterCellAmbush
	}
	return chapterCellActive
}

func selectAttachmentID(attachment uint32, template *chapterTemplate) uint32 {
	switch attachment {
	case chapterAttachBoss:
		return 0
	case chapterAttachEnemy:
		return selectExpeditionFromWeights(template.ExpeditionWeight)
	case chapterAttachElite:
		return selectFirst(template.EliteExpeditions)
	case chapterAttachAmbush:
		if id := selectFirst(template.AmbushExpeditions); id != 0 {
			return id
		}
		return selectExpeditionFromWeights(template.ExpeditionWeight)
	case chapterAttachChampion, chapterAttachBombEnemy, chapterAttachTorpedoEnemy:
		if id := selectFirst(template.GuarderExpeditions); id != 0 {
			return id
		}
		return selectExpeditionFromWeights(template.ExpeditionWeight)
	default:
		return 0
	}
}

func selectExpeditionFromWeights(weights [][]any) uint32 {
	for _, entry := range weights {
		if len(entry) == 0 {
			continue
		}
		id, err := parseUint32(entry[0])
		if err == nil && id != 0 {
			return id
		}
	}
	return 0
}

func selectFirst(values []uint32) uint32 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}

func selectSpawnPositions(grids []chapterGrid, attachment uint32) []chapterPos {
	positions := []chapterPos{}
	for _, grid := range grids {
		if grid.Attachment == attachment {
			positions = append(positions, chapterPos{Row: grid.Row, Column: grid.Column})
		}
	}
	if len(positions) == 0 && len(grids) > 0 {
		positions = append(positions, chapterPos{Row: grids[0].Row, Column: grids[0].Column})
	}
	return positions
}

func chooseSpawn(spawns []chapterPos, index int) chapterPos {
	if len(spawns) == 0 {
		return chapterPos{Row: 1, Column: 1}
	}
	if index < len(spawns) {
		return spawns[index]
	}
	return spawns[0]
}

func buildPos(pos chapterPos) *protobuf.CHAPTERCELLPOS_P13 {
	return &protobuf.CHAPTERCELLPOS_P13{
		Row:    proto.Uint32(pos.Row),
		Column: proto.Uint32(pos.Column),
	}
}

func parseChapterGrids(raw [][]any) ([]chapterGrid, error) {
	grids := make([]chapterGrid, 0, len(raw))
	for _, entry := range raw {
		if len(entry) < 4 {
			return nil, fmt.Errorf("invalid grid entry")
		}
		row, err := parseUint32(entry[0])
		if err != nil {
			return nil, err
		}
		column, err := parseUint32(entry[1])
		if err != nil {
			return nil, err
		}
		walkable, err := parseBool(entry[2])
		if err != nil {
			return nil, err
		}
		attachment, err := parseUint32(entry[3])
		if err != nil {
			return nil, err
		}
		grids = append(grids, chapterGrid{
			Row:        row,
			Column:     column,
			Walkable:   walkable,
			Attachment: attachment,
		})
	}
	return grids, nil
}

func parseUint32(value any) (uint32, error) {
	switch typed := value.(type) {
	case float64:
		return uint32(typed), nil
	case int:
		return uint32(typed), nil
	case int64:
		return uint32(typed), nil
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, err
		}
		return uint32(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported number")
	}
}

func parseBool(value any) (bool, error) {
	switch typed := value.(type) {
	case bool:
		return typed, nil
	case float64:
		return typed != 0, nil
	case int:
		return typed != 0, nil
	case int64:
		return typed != 0, nil
	default:
		return false, fmt.Errorf("unsupported bool")
	}
}
