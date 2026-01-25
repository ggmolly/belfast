package handlers

import (
	"errors"
	"time"

	"github.com/kataras/iris/v12"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

// PlayerChapterState godoc
// @Summary     Get player chapter state
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerChapterStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-state [get]
func (handler *PlayerHandler) PlayerChapterState(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	state, err := orm.GetChapterState(orm.GormDB, commanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "chapter state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load chapter state", nil))
		return
	}
	current, err := decodeChapterState(state.State)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to decode chapter state", nil))
		return
	}
	payload := types.PlayerChapterStateResponse{
		ChapterID: state.ChapterID,
		UpdatedAt: state.UpdatedAt,
		State:     current,
	}
	_ = ctx.JSON(response.Success(payload))
}

// SearchPlayerChapterStates godoc
// @Summary     Search player chapter states
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Param       chapter_id  query  int  false  "Filter by chapter ID"
// @Param       updated_since  query  string  false  "Filter by updated_at >= RFC3339"
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  PlayerChapterStateListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-state/search [get]
func (handler *PlayerHandler) SearchPlayerChapterStates(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	meta, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	query := orm.GormDB.Model(&orm.ChapterState{}).Where("commander_id = ?", commanderID)
	chapterIDParam := ctx.URLParamDefault("chapter_id", "")
	if chapterIDParam != "" {
		chapterID, err := parsePathUint32(chapterIDParam, "chapter_id")
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
			return
		}
		query = query.Where("chapter_id = ?", chapterID)
	}
	updatedSince := ctx.URLParamDefault("updated_since", "")
	if updatedSince != "" {
		parsed, err := time.Parse(time.RFC3339, updatedSince)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid updated_since", nil))
			return
		}
		query = query.Where("updated_at >= ?", uint32(parsed.Unix()))
	}
	if err := query.Count(&meta.Total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to count chapter states", nil))
		return
	}
	var states []orm.ChapterState
	query = query.Order("updated_at desc")
	query = orm.ApplyPagination(query, meta.Offset, meta.Limit)
	if err := query.Find(&states).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load chapter states", nil))
		return
	}
	entries := make([]types.PlayerChapterStateResponse, 0, len(states))
	for _, state := range states {
		decoded, err := decodeChapterState(state.State)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to decode chapter state", nil))
			return
		}
		entries = append(entries, types.PlayerChapterStateResponse{
			ChapterID: state.ChapterID,
			UpdatedAt: state.UpdatedAt,
			State:     decoded,
		})
	}
	_ = ctx.JSON(response.Success(types.PlayerChapterStateListResponse{States: entries, Meta: meta}))
}

// CreatePlayerChapterState godoc
// @Summary     Create player chapter state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerChapterStateCreateRequest  true  "Chapter state payload"
// @Success     200  {object}  PlayerChapterStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-state [post]
func (handler *PlayerHandler) CreatePlayerChapterState(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerChapterStateCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "validation failed", validationErrors(err)))
		return
	}
	if req.State.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "state id required", nil))
		return
	}
	protoState := encodeChapterState(req.State)
	stateBytes, err := proto.Marshal(protoState)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to encode chapter state", nil))
		return
	}
	state := orm.ChapterState{
		CommanderID: commanderID,
		ChapterID:   req.State.ID,
		State:       stateBytes,
	}
	if err := orm.UpsertChapterState(orm.GormDB, &state); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to store chapter state", nil))
		return
	}
	payload := types.PlayerChapterStateResponse{
		ChapterID: state.ChapterID,
		UpdatedAt: state.UpdatedAt,
		State:     req.State,
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerChapterState godoc
// @Summary     Update player chapter state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerChapterStateUpdateRequest  true  "Chapter state payload"
// @Success     200  {object}  PlayerChapterStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-state [patch]
func (handler *PlayerHandler) UpdatePlayerChapterState(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerChapterStateUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if err := handler.Validate.Struct(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "validation failed", validationErrors(err)))
		return
	}
	if req.State.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "state id required", nil))
		return
	}
	protoState := encodeChapterState(req.State)
	stateBytes, err := proto.Marshal(protoState)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to encode chapter state", nil))
		return
	}
	state := orm.ChapterState{
		CommanderID: commanderID,
		ChapterID:   req.State.ID,
		State:       stateBytes,
	}
	if err := orm.UpsertChapterState(orm.GormDB, &state); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to store chapter state", nil))
		return
	}
	payload := types.PlayerChapterStateResponse{
		ChapterID: state.ChapterID,
		UpdatedAt: state.UpdatedAt,
		State:     req.State,
	}
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerChapterState godoc
// @Summary     Delete player chapter state
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-state [delete]
func (handler *PlayerHandler) DeletePlayerChapterState(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	if err := orm.DeleteChapterState(orm.GormDB, commanderID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete chapter state", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

func decodeChapterState(raw []byte) (types.ChapterState, error) {
	var current protobuf.CURRENTCHAPTERINFO
	if err := proto.Unmarshal(raw, &current); err != nil {
		return types.ChapterState{}, err
	}
	return buildChapterStateDTO(&current), nil
}

func encodeChapterState(state types.ChapterState) *protobuf.CURRENTCHAPTERINFO {
	return &protobuf.CURRENTCHAPTERINFO{
		Id:                    proto.Uint32(state.ID),
		Time:                  proto.Uint32(state.Time),
		CellList:              buildChapterCellsProto(state.CellList),
		MainGroupList:         buildChapterGroupsProto(state.MainGroupList),
		AiList:                buildChapterCellsProto(state.AiList),
		EscortList:            buildChapterCellsProto(state.EscortList),
		Round:                 proto.Uint32(state.Round),
		IsSubmarineAutoAttack: proto.Uint32(state.IsSubmarineAutoAttack),
		OperationBuff:         state.OperationBuff,
		ModelActCount:         proto.Uint32(state.ModelActCount),
		BuffList:              state.BuffList,
		LoopFlag:              proto.Uint32(state.LoopFlag),
		ExtraFlagList:         state.ExtraFlagList,
		CellFlagList:          buildChapterCellFlagsProto(state.CellFlagList),
		ChapterHp:             proto.Uint32(state.ChapterHp),
		ChapterStrategyList:   buildChapterStrategiesProto(state.ChapterStrategyList),
		KillCount:             proto.Uint32(state.KillCount),
		InitShipCount:         proto.Uint32(state.InitShipCount),
		ContinuousKillCount:   proto.Uint32(state.ContinuousKillCount),
		BattleStatistics:      buildChapterStrategiesProto(state.BattleStatistics),
		FleetDuties:           buildChapterFleetDutiesProto(state.FleetDuties),
		MoveStepCount:         proto.Uint32(state.MoveStepCount),
		SubmarineGroupList:    buildChapterGroupsProto(state.SubmarineGroupList),
		SupportGroupList:      buildChapterGroupsProto(state.SupportGroupList),
	}
}

func buildChapterStateDTO(current *protobuf.CURRENTCHAPTERINFO) types.ChapterState {
	return types.ChapterState{
		ID:                    current.GetId(),
		Time:                  current.GetTime(),
		CellList:              buildChapterCellsDTO(current.GetCellList()),
		MainGroupList:         buildChapterGroupsDTO(current.GetMainGroupList()),
		AiList:                buildChapterCellsDTO(current.GetAiList()),
		EscortList:            buildChapterCellsDTO(current.GetEscortList()),
		Round:                 current.GetRound(),
		IsSubmarineAutoAttack: current.GetIsSubmarineAutoAttack(),
		OperationBuff:         current.GetOperationBuff(),
		ModelActCount:         current.GetModelActCount(),
		BuffList:              current.GetBuffList(),
		LoopFlag:              current.GetLoopFlag(),
		ExtraFlagList:         current.GetExtraFlagList(),
		CellFlagList:          buildChapterCellFlagsDTO(current.GetCellFlagList()),
		ChapterHp:             current.GetChapterHp(),
		ChapterStrategyList:   buildChapterStrategiesDTO(current.GetChapterStrategyList()),
		KillCount:             current.GetKillCount(),
		InitShipCount:         current.GetInitShipCount(),
		ContinuousKillCount:   current.GetContinuousKillCount(),
		BattleStatistics:      buildChapterStrategiesDTO(current.GetBattleStatistics()),
		FleetDuties:           buildChapterFleetDutiesDTO(current.GetFleetDuties()),
		MoveStepCount:         current.GetMoveStepCount(),
		SubmarineGroupList:    buildChapterGroupsDTO(current.GetSubmarineGroupList()),
		SupportGroupList:      buildChapterGroupsDTO(current.GetSupportGroupList()),
	}
}

func buildChapterCellsDTO(cells []*protobuf.CHAPTERCELLINFO_P13) []types.ChapterCellInfo {
	result := make([]types.ChapterCellInfo, 0, len(cells))
	for _, cell := range cells {
		entry := types.ChapterCellInfo{
			Pos:      buildChapterPosDTO(cell.GetPos()),
			ItemType: cell.GetItemType(),
			ExtraID:  cell.GetExtraId(),
		}
		if cell.ItemId != nil {
			value := cell.GetItemId()
			entry.ItemID = &value
		}
		if cell.ItemFlag != nil {
			value := cell.GetItemFlag()
			entry.ItemFlag = &value
		}
		if cell.ItemData != nil {
			value := cell.GetItemData()
			entry.ItemData = &value
		}
		result = append(result, entry)
	}
	return result
}

func buildChapterGroupsDTO(groups []*protobuf.GROUPINCHAPTER_P13) []types.ChapterGroup {
	result := make([]types.ChapterGroup, 0, len(groups))
	for _, group := range groups {
		entry := types.ChapterGroup{
			ID:               group.GetId(),
			ShipList:         buildChapterShipsDTO(group.GetShipList()),
			Pos:              buildChapterPosDTO(group.GetPos()),
			StepCount:        group.GetStepCount(),
			BoxStrategyList:  buildChapterStrategiesDTO(group.GetBoxStrategyList()),
			ShipStrategyList: buildChapterStrategiesDTO(group.GetShipStrategyList()),
			StrategyIds:      group.GetStrategyIds(),
			Bullet:           group.GetBullet(),
			StartPos:         buildChapterPosDTO(group.GetStartPos()),
			CommanderList:    buildChapterCommandersDTO(group.GetCommanderList()),
			MoveStepDown:     group.GetMoveStepDown(),
			KillCount:        group.GetKillCount(),
			FleetId:          group.GetFleetId(),
			VisionLv:         group.GetVisionLv(),
		}
		result = append(result, entry)
	}
	return result
}

func buildChapterShipsDTO(ships []*protobuf.SHIPINCHAPTER_P13) []types.ChapterShip {
	result := make([]types.ChapterShip, 0, len(ships))
	for _, ship := range ships {
		result = append(result, types.ChapterShip{
			ID:     ship.GetId(),
			HpRant: ship.GetHpRant(),
		})
	}
	return result
}

func buildChapterCommandersDTO(commanders []*protobuf.COMMANDERSINFO) []types.ChapterCommander {
	result := make([]types.ChapterCommander, 0, len(commanders))
	for _, commander := range commanders {
		result = append(result, types.ChapterCommander{
			Pos: commander.GetPos(),
			ID:  commander.GetId(),
		})
	}
	return result
}

func buildChapterStrategiesDTO(strategies []*protobuf.STRATEGYINFO_P13) []types.ChapterStrategy {
	result := make([]types.ChapterStrategy, 0, len(strategies))
	for _, strategy := range strategies {
		result = append(result, types.ChapterStrategy{
			ID:    strategy.GetId(),
			Count: strategy.GetCount(),
		})
	}
	return result
}

func buildChapterCellFlagsDTO(flags []*protobuf.CELLFLAG) []types.ChapterCellFlag {
	result := make([]types.ChapterCellFlag, 0, len(flags))
	for _, flag := range flags {
		result = append(result, types.ChapterCellFlag{
			Pos:      buildChapterPosDTO(flag.GetPos()),
			FlagList: flag.GetFlagList(),
		})
	}
	return result
}

func buildChapterFleetDutiesDTO(duties []*protobuf.FLEETDUTYKEYVALUEPAIR) []types.ChapterFleetDuty {
	result := make([]types.ChapterFleetDuty, 0, len(duties))
	for _, duty := range duties {
		result = append(result, types.ChapterFleetDuty{
			Key:   duty.GetKey(),
			Value: duty.GetValue(),
		})
	}
	return result
}

func buildChapterPosDTO(pos *protobuf.CHAPTERCELLPOS_P13) types.ChapterCellPos {
	if pos == nil {
		return types.ChapterCellPos{}
	}
	return types.ChapterCellPos{
		Row:    pos.GetRow(),
		Column: pos.GetColumn(),
	}
}

func buildChapterCellsProto(cells []types.ChapterCellInfo) []*protobuf.CHAPTERCELLINFO_P13 {
	result := make([]*protobuf.CHAPTERCELLINFO_P13, 0, len(cells))
	for _, cell := range cells {
		entry := &protobuf.CHAPTERCELLINFO_P13{
			Pos:      buildChapterPosProto(cell.Pos),
			ItemType: proto.Uint32(cell.ItemType),
			ExtraId:  cell.ExtraID,
		}
		if cell.ItemID != nil {
			entry.ItemId = proto.Uint32(*cell.ItemID)
		}
		if cell.ItemFlag != nil {
			entry.ItemFlag = proto.Uint32(*cell.ItemFlag)
		}
		if cell.ItemData != nil {
			entry.ItemData = proto.Uint32(*cell.ItemData)
		}
		result = append(result, entry)
	}
	return result
}

func buildChapterGroupsProto(groups []types.ChapterGroup) []*protobuf.GROUPINCHAPTER_P13 {
	result := make([]*protobuf.GROUPINCHAPTER_P13, 0, len(groups))
	for _, group := range groups {
		result = append(result, &protobuf.GROUPINCHAPTER_P13{
			Id:               proto.Uint32(group.ID),
			ShipList:         buildChapterShipsProto(group.ShipList),
			Pos:              buildChapterPosProto(group.Pos),
			StepCount:        proto.Uint32(group.StepCount),
			BoxStrategyList:  buildChapterStrategiesProto(group.BoxStrategyList),
			ShipStrategyList: buildChapterStrategiesProto(group.ShipStrategyList),
			StrategyIds:      group.StrategyIds,
			Bullet:           proto.Uint32(group.Bullet),
			StartPos:         buildChapterPosProto(group.StartPos),
			CommanderList:    buildChapterCommandersProto(group.CommanderList),
			MoveStepDown:     proto.Uint32(group.MoveStepDown),
			KillCount:        proto.Uint32(group.KillCount),
			FleetId:          proto.Uint32(group.FleetId),
			VisionLv:         proto.Uint32(group.VisionLv),
		})
	}
	return result
}

func buildChapterShipsProto(ships []types.ChapterShip) []*protobuf.SHIPINCHAPTER_P13 {
	result := make([]*protobuf.SHIPINCHAPTER_P13, 0, len(ships))
	for _, ship := range ships {
		result = append(result, &protobuf.SHIPINCHAPTER_P13{
			Id:     proto.Uint32(ship.ID),
			HpRant: proto.Uint32(ship.HpRant),
		})
	}
	return result
}

func buildChapterCommandersProto(commanders []types.ChapterCommander) []*protobuf.COMMANDERSINFO {
	result := make([]*protobuf.COMMANDERSINFO, 0, len(commanders))
	for _, commander := range commanders {
		result = append(result, &protobuf.COMMANDERSINFO{
			Pos: proto.Uint32(commander.Pos),
			Id:  proto.Uint32(commander.ID),
		})
	}
	return result
}

func buildChapterStrategiesProto(strategies []types.ChapterStrategy) []*protobuf.STRATEGYINFO_P13 {
	result := make([]*protobuf.STRATEGYINFO_P13, 0, len(strategies))
	for _, strategy := range strategies {
		result = append(result, &protobuf.STRATEGYINFO_P13{
			Id:    proto.Uint32(strategy.ID),
			Count: proto.Uint32(strategy.Count),
		})
	}
	return result
}

func buildChapterCellFlagsProto(flags []types.ChapterCellFlag) []*protobuf.CELLFLAG {
	result := make([]*protobuf.CELLFLAG, 0, len(flags))
	for _, flag := range flags {
		result = append(result, &protobuf.CELLFLAG{
			Pos:      buildChapterPosProto(flag.Pos),
			FlagList: flag.FlagList,
		})
	}
	return result
}

func buildChapterFleetDutiesProto(duties []types.ChapterFleetDuty) []*protobuf.FLEETDUTYKEYVALUEPAIR {
	result := make([]*protobuf.FLEETDUTYKEYVALUEPAIR, 0, len(duties))
	for _, duty := range duties {
		result = append(result, &protobuf.FLEETDUTYKEYVALUEPAIR{
			Key:   proto.Uint32(duty.Key),
			Value: proto.Uint32(duty.Value),
		})
	}
	return result
}

func buildChapterPosProto(pos types.ChapterCellPos) *protobuf.CHAPTERCELLPOS_P13 {
	return &protobuf.CHAPTERCELLPOS_P13{
		Row:    proto.Uint32(pos.Row),
		Column: proto.Uint32(pos.Column),
	}
}
