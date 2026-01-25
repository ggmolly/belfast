package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

// PlayerRemasterState godoc
// @Summary     Get player remaster state
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerRemasterStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/remaster [get]
func (handler *PlayerHandler) PlayerRemasterState(ctx iris.Context) {
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
	state, err := orm.GetOrCreateRemasterState(orm.GormDB, commanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load remaster state", nil))
		return
	}
	payload := types.PlayerRemasterStateResponse{
		TicketCount:      state.TicketCount,
		DailyCount:       state.DailyCount,
		LastDailyResetAt: state.LastDailyResetAt.UTC(),
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerRemasterState godoc
// @Summary     Update player remaster state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerRemasterStateUpdateRequest  true  "Remaster state update"
// @Success     200  {object}  PlayerRemasterStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/remaster [patch]
func (handler *PlayerHandler) UpdatePlayerRemasterState(ctx iris.Context) {
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
	var req types.PlayerRemasterStateUpdateRequest
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
	state, err := orm.GetOrCreateRemasterState(orm.GormDB, commanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load remaster state", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.TicketCount != nil {
		state.TicketCount = *req.TicketCount
		updates["ticket_count"] = *req.TicketCount
	}
	if req.DailyCount != nil {
		state.DailyCount = *req.DailyCount
		updates["daily_count"] = *req.DailyCount
	}
	if req.LastDailyResetAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.LastDailyResetAt)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid last_daily_reset_at", nil))
			return
		}
		state.LastDailyResetAt = parsed
		updates["last_daily_reset_at"] = parsed
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(state).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update remaster state", nil))
		return
	}
	payload := types.PlayerRemasterStateResponse{
		TicketCount:      state.TicketCount,
		DailyCount:       state.DailyCount,
		LastDailyResetAt: state.LastDailyResetAt.UTC(),
	}
	_ = ctx.JSON(response.Success(payload))
}

// PlayerRemasterProgress godoc
// @Summary     Get player remaster progress
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       chapter_id  query  int  false  "Filter by chapter ID"
// @Param       received  query  bool  false  "Filter by received"
// @Success     200  {object}  PlayerRemasterProgressResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/remaster/progress [get]
func (handler *PlayerHandler) PlayerRemasterProgress(ctx iris.Context) {
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
	query := orm.GormDB.Where("commander_id = ?", commanderID)
	chapterIDParam := strings.TrimSpace(ctx.URLParam("chapter_id"))
	if chapterIDParam != "" {
		chapterID, err := parsePathUint32(chapterIDParam, "chapter_id")
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
			return
		}
		query = query.Where("chapter_id = ?", chapterID)
	}
	receivedParam := strings.TrimSpace(ctx.URLParam("received"))
	if receivedParam != "" {
		received, err := parseOptionalBool(receivedParam)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
			return
		}
		query = query.Where("received = ?", received)
	}
	var progress []orm.RemasterProgress
	if err := query.Order("chapter_id asc, pos asc").Find(&progress).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load remaster progress", nil))
		return
	}
	entries := make([]types.PlayerRemasterProgressEntry, 0, len(progress))
	for _, entry := range progress {
		entries = append(entries, types.PlayerRemasterProgressEntry{
			ChapterID: entry.ChapterID,
			Pos:       entry.Pos,
			Count:     entry.Count,
			Received:  entry.Received,
			UpdatedAt: entry.UpdatedAt.UTC(),
		})
	}
	payload := types.PlayerRemasterProgressResponse{Progress: entries}
	_ = ctx.JSON(response.Success(payload))
}

// UpsertPlayerRemasterProgress godoc
// @Summary     Create or update player remaster progress
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerRemasterProgressCreateRequest  true  "Remaster progress"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/remaster/progress [post]
func (handler *PlayerHandler) UpsertPlayerRemasterProgress(ctx iris.Context) {
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
	var req types.PlayerRemasterProgressCreateRequest
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
	entry := orm.RemasterProgress{
		CommanderID: commanderID,
		ChapterID:   req.ChapterID,
		Pos:         req.Pos,
		Count:       req.Count,
		Received:    false,
	}
	if req.Received != nil {
		entry.Received = *req.Received
	} else {
		existing, err := orm.GetRemasterProgress(orm.GormDB, commanderID, req.ChapterID, req.Pos)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load remaster progress", nil))
			return
		}
		if err == nil {
			entry.Received = existing.Received
		}
	}
	if err := orm.UpsertRemasterProgress(orm.GormDB, &entry); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to upsert remaster progress", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdatePlayerRemasterProgress godoc
// @Summary     Update player remaster progress
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       chapter_id  path  int  true  "Chapter ID"
// @Param       pos  path  int  true  "Reward position"
// @Param       payload  body  types.PlayerRemasterProgressUpdateRequest  true  "Remaster progress update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/remaster/progress/{chapter_id}/{pos} [patch]
func (handler *PlayerHandler) UpdatePlayerRemasterProgress(ctx iris.Context) {
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
	chapterID, err := parsePathUint32(ctx.Params().Get("chapter_id"), "chapter_id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	pos, err := parsePathUint32(ctx.Params().Get("pos"), "pos")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.PlayerRemasterProgressUpdateRequest
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
	progress, err := orm.GetRemasterProgress(orm.GormDB, commanderID, chapterID, pos)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "remaster progress not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load remaster progress", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.Count != nil {
		updates["count"] = *req.Count
	}
	if req.Received != nil {
		updates["received"] = *req.Received
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(progress).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update remaster progress", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerRemasterProgress godoc
// @Summary     Delete player remaster progress
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       chapter_id  path  int  true  "Chapter ID"
// @Param       pos  path  int  true  "Reward position"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/remaster/progress/{chapter_id}/{pos} [delete]
func (handler *PlayerHandler) DeletePlayerRemasterProgress(ctx iris.Context) {
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
	chapterID, err := parsePathUint32(ctx.Params().Get("chapter_id"), "chapter_id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	pos, err := parsePathUint32(ctx.Params().Get("pos"), "pos")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.DeleteRemasterProgress(orm.GormDB, commanderID, chapterID, pos); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete remaster progress", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}
