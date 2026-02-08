package answer

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"sync"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/rng"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

const (
	chapterOpRetreat    = 0
	chapterOpMove       = 1
	chapterOpAmbush     = 4
	chapterOpEnemyRound = 8
	chapterOpRequest    = 49
)

const (
	chapterChanceBase = 10000
)

var chapterAmbushRand = rng.NewLockedRand()

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
		mapUpdate := []*protobuf.CHAPTERCELLINFO_P13{}
		if ambushCell := maybeTriggerChapterAmbush(template, &current, group, end, client); ambushCell != nil {
			mapUpdate = append(mapUpdate, ambushCell)
		}
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
			Result:    proto.Uint32(0),
			MovePath:  movePath,
			MapUpdate: mapUpdate,
		}
		return client.SendMessage(13104, &response)
	case chapterOpAmbush:
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
		pos := chapterPos{Row: group.GetPos().GetRow(), Column: group.GetPos().GetColumn()}
		idx, cell := findChapterCellAt(&current, pos)
		if cell == nil || cell.GetItemType() != chapterAttachAmbush {
			response := protobuf.SC_13104{Result: proto.Uint32(1)}
			return client.SendMessage(13104, &response)
		}
		mapUpdate := []*protobuf.CHAPTERCELLINFO_P13{}
		arg := payload.GetActArg_1()
		if arg == 1 {
			threshold := calculateAmbushDodgeThreshold(template, group, pos, client)
			if threshold > 0 && chapterAmbushRand.Uint32N(chapterChanceBase) < threshold {
				current.CellList = append(current.CellList[:idx], current.CellList[idx+1:]...)
			} else {
				ensureAmbushCellHasExpedition(cell, template)
				cell.ItemFlag = proto.Uint32(chapterCellActive)
				mapUpdate = append(mapUpdate, cell)
			}
		} else {
			ensureAmbushCellHasExpedition(cell, template)
			cell.ItemFlag = proto.Uint32(chapterCellActive)
			mapUpdate = append(mapUpdate, cell)
		}
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
			MapUpdate:    mapUpdate,
			ShipUpdate:   collectChapterShips(&current),
			AiList:       current.GetAiList(),
			BuffList:     current.GetBuffList(),
			CellFlagList: current.GetCellFlagList(),
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

func maybeTriggerChapterAmbush(template *chapterTemplate, current *protobuf.CURRENTCHAPTERINFO, group *protobuf.GROUPINCHAPTER_P13, end chapterPos, client *connection.Client) *protobuf.CHAPTERCELLINFO_P13 {
	if template == nil || current == nil || group == nil || client == nil {
		return nil
	}
	if template.IsAmbush == 0 {
		return nil
	}
	if end.Row == 0 || end.Column == 0 {
		return nil
	}
	if _, cell := findChapterCellAt(current, end); cell != nil {
		if cell.GetItemType() != chapterAttachBorn && cell.GetItemType() != chapterAttachBornSub {
			return nil
		}
	}
	threshold := calculateAmbushTriggerThreshold(template, group, end, client)
	if threshold == 0 {
		return nil
	}
	if chapterAmbushRand.Uint32N(chapterChanceBase) >= threshold {
		return nil
	}
	expeditionID := selectFirst(template.AmbushExpeditions)
	if expeditionID == 0 {
		expeditionID = selectExpeditionFromWeights(template.ExpeditionWeight)
	}
	if expeditionID == 0 {
		return nil
	}
	cell := &protobuf.CHAPTERCELLINFO_P13{
		Pos:      buildPos(end),
		ItemType: proto.Uint32(chapterAttachAmbush),
		ItemId:   proto.Uint32(expeditionID),
		ItemFlag: proto.Uint32(chapterCellAmbush),
		ItemData: proto.Uint32(0),
	}
	upsertChapterCell(current, cell)
	return cell
}

func calculateAmbushTriggerThreshold(template *chapterTemplate, group *protobuf.GROUPINCHAPTER_P13, pos chapterPos, client *connection.Client) uint32 {
	// Mirrors the client formula in AzurLaneLuaScripts/*/model/vo/chapterleveldata.lua
	// chapterleveldata.getAmbushRate().
	if template == nil || group == nil || client == nil {
		return 0
	}
	step := int(group.GetStepCount())
	if step > 0 {
		step--
	}
	inv := float64(template.InvestigationRatio)
	investSums := calculateGroupInvestSums(group, client)
	posExtra, globalExtra := chapterAmbushRatioExtras(template, pos)
	rate := 0.05 + posExtra + globalExtra
	if step > 0 {
		denom := inv + investSums
		if denom > 0 {
			rate += (inv / denom) / 4 * float64(step)
		}
	}
	if posExtra == 0 {
		rate -= calculateFleetEquipAmbushRateReduce(group, client)
	}
	rate = clampChance(rate)
	return uint32(rate * chapterChanceBase)
}

func calculateAmbushDodgeThreshold(template *chapterTemplate, group *protobuf.GROUPINCHAPTER_P13, pos chapterPos, client *connection.Client) uint32 {
	// Mirrors the client formula in AzurLaneLuaScripts/*/model/vo/chapterleveldata.lua
	// chapterleveldata.getAmbushDodge().
	if template == nil || group == nil || client == nil {
		return 0
	}
	avoid := float64(template.AvoidRatio)
	if avoid <= 0 {
		return chapterChanceBase
	}
	dodgeSums := calculateGroupDodgeSums(group, client)
	if dodgeSums <= 0 {
		return 0
	}
	chance := dodgeSums / (dodgeSums + avoid)
	posExtra, _ := chapterAmbushRatioExtras(template, pos)
	if posExtra == 0 {
		chance += calculateFleetEquipDodgeRateUp(group, client)
	}
	chance = clampChance(chance)
	return uint32(chance * chapterChanceBase)
}

func calculateGroupInvestSums(group *protobuf.GROUPINCHAPTER_P13, client *connection.Client) float64 {
	sum := calculateGroupInvestSumBase(group, client)
	if sum <= 0 {
		return 0
	}
	return math.Pow(sum, 2.0/3.0)
}

func calculateGroupDodgeSums(group *protobuf.GROUPINCHAPTER_P13, client *connection.Client) float64 {
	sum := calculateGroupDodgeSumBase(group, client)
	if sum <= 0 {
		return 0
	}
	return math.Pow(sum, 2.0/3.0)
}

func calculateGroupInvestSumBase(group *protobuf.GROUPINCHAPTER_P13, client *connection.Client) float64 {
	if group == nil || client == nil || client.Commander == nil {
		return 0
	}
	if client.Commander.OwnedShipsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0
		}
	}
	var sum float64
	for _, ship := range group.GetShipList() {
		id := ship.GetId()
		if id == 0 {
			continue
		}
		owned, ok := client.Commander.OwnedShipsMap[id]
		if !ok {
			continue
		}
		air, dodge := calculateShipPropertiesAirDodge(owned, client.Commander.CommanderID)
		sum += air + dodge
	}
	return sum
}

func calculateGroupDodgeSumBase(group *protobuf.GROUPINCHAPTER_P13, client *connection.Client) float64 {
	if group == nil || client == nil || client.Commander == nil {
		return 0
	}
	if client.Commander.OwnedShipsMap == nil {
		if err := client.Commander.Load(); err != nil {
			return 0
		}
	}
	var sum float64
	for _, ship := range group.GetShipList() {
		id := ship.GetId()
		if id == 0 {
			continue
		}
		owned, ok := client.Commander.OwnedShipsMap[id]
		if !ok {
			continue
		}
		_, dodge := calculateShipPropertiesAirDodge(owned, client.Commander.CommanderID)
		sum += dodge
	}
	return sum
}

func ensureAmbushCellHasExpedition(cell *protobuf.CHAPTERCELLINFO_P13, template *chapterTemplate) {
	if cell == nil || template == nil {
		return
	}
	if cell.GetItemId() != 0 {
		return
	}
	expeditionID := selectFirst(template.AmbushExpeditions)
	if expeditionID == 0 {
		expeditionID = selectExpeditionFromWeights(template.ExpeditionWeight)
	}
	if expeditionID == 0 {
		return
	}
	cell.ItemId = proto.Uint32(expeditionID)
}

func findChapterCellAt(current *protobuf.CURRENTCHAPTERINFO, pos chapterPos) (int, *protobuf.CHAPTERCELLINFO_P13) {
	if current == nil {
		return -1, nil
	}
	for i, cell := range current.GetCellList() {
		p := cell.GetPos()
		if p == nil {
			continue
		}
		if p.GetRow() == pos.Row && p.GetColumn() == pos.Column {
			return i, cell
		}
	}
	return -1, nil
}

func upsertChapterCell(current *protobuf.CURRENTCHAPTERINFO, cell *protobuf.CHAPTERCELLINFO_P13) {
	if current == nil || cell == nil || cell.GetPos() == nil {
		return
	}
	pos := chapterPos{Row: cell.GetPos().GetRow(), Column: cell.GetPos().GetColumn()}
	idx, _ := findChapterCellAt(current, pos)
	if idx >= 0 {
		current.CellList[idx] = cell
		return
	}
	current.CellList = append(current.CellList, cell)
}

const (
	shipAttrIndexAir   = 4
	shipAttrIndexDodge = 8
)

// The ship stats JSON encodes attrs/attrs_growth as an array ordered by the
// client AttributeType list (see AzurLaneLuaScripts/*/model/vo/ship.lua
// Ship.PROPERTIES). We only need Air and Dodge for chapter ambush calculations.

type shipDataStatisticsEntry struct {
	ID               uint32    `json:"id"`
	Attrs            []float64 `json:"attrs"`
	AttrsGrowth      []float64 `json:"attrs_growth"`
	AttrsGrowthExtra []float64 `json:"attrs_growth_extra"`
}

type equipDataStatisticsEntry struct {
	ID              uint32         `json:"id"`
	Attribute1      *string        `json:"attribute_1"`
	Value1          any            `json:"value_1"`
	Attribute2      *string        `json:"attribute_2"`
	Value2          any            `json:"value_2"`
	Attribute3      *string        `json:"attribute_3"`
	Value3          any            `json:"value_3"`
	EquipParameters map[string]any `json:"equip_parameters"`
}

type gamesetKeyValueEntry struct {
	KeyValue uint32 `json:"key_value"`
}

type intimacyTemplateEntry struct {
	ID         uint32 `json:"id"`
	LowerBound uint32 `json:"lower_bound"`
	UpperBound uint32 `json:"upper_bound"`
	AttrBonus  uint32 `json:"attr_bonus"`
}

type transformDataEntry struct {
	ID     uint32               `json:"id"`
	Effect []map[string]float64 `json:"effect"`
}

var (
	extraAttrLevelLimitOnce sync.Once
	extraAttrLevelLimit     uint32 = 100

	intimacyTemplateOnce sync.Once
	intimacyTemplates    []intimacyTemplateEntry
)

func getExtraAttrLevelLimit() uint32 {
	extraAttrLevelLimitOnce.Do(func() {
		entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/gameset.json", "extra_attr_level_limit")
		if err != nil || entry == nil {
			return
		}
		var config gamesetKeyValueEntry
		if err := json.Unmarshal(entry.Data, &config); err != nil {
			return
		}
		if config.KeyValue > 0 {
			extraAttrLevelLimit = config.KeyValue
		}
	})
	return extraAttrLevelLimit
}

func loadIntimacyTemplates() {
	entries, err := orm.ListConfigEntries(orm.GormDB, "ShareCfg/intimacy_template.json")
	if err != nil {
		return
	}
	templates := make([]intimacyTemplateEntry, 0, len(entries))
	for _, entry := range entries {
		var tpl intimacyTemplateEntry
		if err := json.Unmarshal(entry.Data, &tpl); err != nil {
			continue
		}
		templates = append(templates, tpl)
	}
	intimacyTemplates = templates
}

func intimacyAttrBonusRate(intimacy uint32) float64 {
	intimacyTemplateOnce.Do(loadIntimacyTemplates)
	for _, tpl := range intimacyTemplates {
		if intimacy >= tpl.LowerBound && intimacy <= tpl.UpperBound {
			return float64(tpl.AttrBonus) / 10000
		}
	}
	return 0
}

func calcFloor(value float64) float64 {
	return math.Floor(value + 1e-9)
}

func clampChance(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func chapterAmbushRatioExtras(template *chapterTemplate, pos chapterPos) (float64, float64) {
	if template == nil {
		return 0, 0
	}
	var posExtra float64
	var globalExtra float64
	for _, entry := range template.AmbushRatioExtra {
		switch len(entry) {
		case 1:
			globalExtra = float64(entry[0]) / chapterChanceBase
		default:
			if len(entry) >= 3 && entry[0] == int32(pos.Row) && entry[1] == int32(pos.Column) {
				posExtra = float64(entry[2]) / chapterChanceBase
			}
		}
	}
	return posExtra, globalExtra
}

func calculateShipPropertiesAirDodge(owned *orm.OwnedShip, ownerID uint32) (float64, float64) {
	// This is the subset of ship:getProperties() required for ambush math:
	// base growth (ship_data_statistics), intimacy bonus, transforms, and
	// flat equipment Air/Dodge additions.
	if owned == nil {
		return 0, 0
	}
	stats := loadShipDataStatistics(owned.ShipID)
	if stats == nil {
		return 0, 0
	}
	extraLimit := getExtraAttrLevelLimit()
	baseAir := shipGrowthForIndex(stats, shipAttrIndexAir, owned.Level, extraLimit)
	baseDodge := shipGrowthForIndex(stats, shipAttrIndexDodge, owned.Level, extraLimit)
	bonusRate := intimacyAttrBonusRate(owned.Intimacy)
	transformAir, transformDodge := shipTransformAdditionsAirDodge(ownerID, owned.ID)
	propAir := baseAir*(1+bonusRate) + transformAir
	propDodge := baseDodge*(1+bonusRate) + transformDodge
	air := calcFloor(propAir)
	dodge := calcFloor(propDodge)
	equipAir, equipDodge := equipmentAttributeAdditions(ownerID, owned.ID)
	air += equipAir
	dodge += equipDodge
	return air, dodge
}

func shipTransformAdditionsAirDodge(ownerID uint32, shipID uint32) (float64, float64) {
	entries, err := orm.ListOwnedShipTransforms(orm.GormDB, ownerID, shipID)
	if err != nil {
		return 0, 0
	}
	var airAdd float64
	var dodgeAdd float64
	for _, owned := range entries {
		cfg := loadTransformData(owned.TransformID)
		if cfg == nil {
			continue
		}
		level := int(owned.Level)
		if level <= 0 {
			continue
		}
		for i := 0; i < level && i < len(cfg.Effect); i++ {
			effect := cfg.Effect[i]
			airAdd += effect["air"]
			dodgeAdd += effect["dodge"]
		}
	}
	return airAdd, dodgeAdd
}

func shipGrowthForIndex(stats *shipDataStatisticsEntry, index int, level uint32, extraLimit uint32) float64 {
	if stats == nil || index < 0 {
		return 0
	}
	if len(stats.Attrs) <= index || len(stats.AttrsGrowth) <= index || len(stats.AttrsGrowthExtra) <= index {
		return 0
	}
	base := stats.Attrs[index]
	if level > 1 {
		base += float64(level-1) * stats.AttrsGrowth[index] / 1000
	}
	if extraLimit > 0 && level > extraLimit {
		base += float64(level-extraLimit) * stats.AttrsGrowthExtra[index] / 1000
	}
	return base
}

func loadShipDataStatistics(templateID uint32) *shipDataStatisticsEntry {
	entry, err := orm.GetConfigEntry(orm.GormDB, "sharecfgdata/ship_data_statistics.json", strconv.FormatUint(uint64(templateID), 10))
	if err != nil || entry == nil {
		return nil
	}
	var stats shipDataStatisticsEntry
	if err := json.Unmarshal(entry.Data, &stats); err != nil {
		return nil
	}
	return &stats
}

func equipmentAttributeAdditions(ownerID uint32, shipID uint32) (float64, float64) {
	entries, err := orm.ListOwnedShipEquipment(orm.GormDB, ownerID, shipID)
	if err != nil {
		return 0, 0
	}
	var airAdd float64
	var dodgeAdd float64
	for _, owned := range entries {
		cfg := loadEquipDataStatistics(owned.EquipID)
		if cfg == nil {
			continue
		}
		airAdd += equipAttributeValue(cfg.Attribute1, cfg.Value1, "air")
		dodgeAdd += equipAttributeValue(cfg.Attribute1, cfg.Value1, "dodge")
		airAdd += equipAttributeValue(cfg.Attribute2, cfg.Value2, "air")
		dodgeAdd += equipAttributeValue(cfg.Attribute2, cfg.Value2, "dodge")
		airAdd += equipAttributeValue(cfg.Attribute3, cfg.Value3, "air")
		dodgeAdd += equipAttributeValue(cfg.Attribute3, cfg.Value3, "dodge")
	}
	return airAdd, dodgeAdd
}

func calculateFleetEquipAmbushRateReduce(group *protobuf.GROUPINCHAPTER_P13, client *connection.Client) float64 {
	if group == nil || client == nil || client.Commander == nil {
		return 0
	}
	ownerID := client.Commander.CommanderID
	var maxExtra float64
	for _, ship := range group.GetShipList() {
		owned, ok := client.Commander.OwnedShipsMap[ship.GetId()]
		if !ok {
			continue
		}
		equipments, err := orm.ListOwnedShipEquipment(orm.GormDB, ownerID, owned.ID)
		if err != nil {
			continue
		}
		for _, eq := range equipments {
			cfg := loadEquipDataStatistics(eq.EquipID)
			if cfg == nil {
				continue
			}
			value := equipParameterRate(cfg.EquipParameters, "ambush_extra")
			if value > maxExtra {
				maxExtra = value
			}
		}
	}
	return maxExtra
}

func calculateFleetEquipDodgeRateUp(group *protobuf.GROUPINCHAPTER_P13, client *connection.Client) float64 {
	if group == nil || client == nil || client.Commander == nil {
		return 0
	}
	ownerID := client.Commander.CommanderID
	var maxExtra float64
	for _, ship := range group.GetShipList() {
		owned, ok := client.Commander.OwnedShipsMap[ship.GetId()]
		if !ok {
			continue
		}
		equipments, err := orm.ListOwnedShipEquipment(orm.GormDB, ownerID, owned.ID)
		if err != nil {
			continue
		}
		for _, eq := range equipments {
			cfg := loadEquipDataStatistics(eq.EquipID)
			if cfg == nil {
				continue
			}
			value := equipParameterRate(cfg.EquipParameters, "avoid_extra")
			if value > maxExtra {
				maxExtra = value
			}
		}
	}
	return maxExtra
}

func loadEquipDataStatistics(equipID uint32) *equipDataStatisticsEntry {
	entry, err := orm.GetConfigEntry(orm.GormDB, "sharecfgdata/equip_data_statistics.json", strconv.FormatUint(uint64(equipID), 10))
	if err != nil || entry == nil {
		return nil
	}
	var stats equipDataStatisticsEntry
	if err := json.Unmarshal(entry.Data, &stats); err != nil {
		return nil
	}
	return &stats
}

func loadTransformData(transformID uint32) *transformDataEntry {
	entry, err := orm.GetConfigEntry(orm.GormDB, "ShareCfg/transform_data_template.json", strconv.FormatUint(uint64(transformID), 10))
	if err != nil || entry == nil {
		return nil
	}
	var stats transformDataEntry
	if err := json.Unmarshal(entry.Data, &stats); err != nil {
		return nil
	}
	return &stats
}

func equipAttributeValue(attr *string, value any, expected string) float64 {
	if attr == nil || *attr != expected {
		return 0
	}
	parsed, ok := parseFloat64(value)
	if !ok {
		return 0
	}
	return parsed
}

func equipParameterRate(params map[string]any, key string) float64 {
	if len(params) == 0 {
		return 0
	}
	value, ok := params[key]
	if !ok {
		return 0
	}
	parsed, ok := parseFloat64(value)
	if !ok {
		return 0
	}
	return parsed / chapterChanceBase
}

func parseFloat64(value any) (float64, bool) {
	// Config JSON isn't type-stable when decoded into `any` (numbers default to
	// float64, but some code paths may produce json.Number or strings).
	// Normalize the common representations into float64 for calculations.
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case json.Number:
		f, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case string:
		f, err := strconv.ParseFloat(typed, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
