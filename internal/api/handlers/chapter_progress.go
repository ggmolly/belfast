package handlers

import (
	"errors"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

// PlayerChapterProgress godoc
// @Summary     Get player chapter progress
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       chapter_id  path  int  true  "Chapter ID"
// @Success     200  {object}  PlayerChapterProgressResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-progress/{chapter_id} [get]
func (handler *PlayerHandler) PlayerChapterProgress(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	chapterID, err := parsePathUint32(ctx.Params().Get("chapter_id"), "chapter_id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	progress, err := orm.GetChapterProgress(orm.GormDB, commanderID, chapterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "chapter progress not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load chapter progress", nil))
		return
	}
	payload := types.PlayerChapterProgressResponse{Progress: buildChapterProgressDTO(progress)}
	_ = ctx.JSON(response.Success(payload))
}

// ListPlayerChapterProgress godoc
// @Summary     List player chapter progress
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  PlayerChapterProgressListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-progress [get]
func (handler *PlayerHandler) ListPlayerChapterProgress(ctx iris.Context) {
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
	query := orm.GormDB.Model(&orm.ChapterProgress{}).Where("commander_id = ?", commanderID)
	if err := query.Count(&meta.Total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to count chapter progress", nil))
		return
	}
	var progress []orm.ChapterProgress
	query = query.Order("chapter_id asc")
	query = orm.ApplyPagination(query, meta.Offset, meta.Limit)
	if err := query.Find(&progress).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load chapter progress", nil))
		return
	}
	entries := make([]types.PlayerChapterProgressResponse, 0, len(progress))
	for _, entry := range progress {
		entries = append(entries, types.PlayerChapterProgressResponse{Progress: buildChapterProgressDTO(&entry)})
	}
	_ = ctx.JSON(response.Success(types.PlayerChapterProgressListResponse{Progress: entries, Meta: meta}))
}

// SearchPlayerChapterProgress godoc
// @Summary     Search player chapter progress
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Param       chapter_id  query  int  false  "Filter by chapter ID"
// @Param       updated_since  query  string  false  "Filter by updated_at >= RFC3339"
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  PlayerChapterProgressListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-progress/search [get]
func (handler *PlayerHandler) SearchPlayerChapterProgress(ctx iris.Context) {
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
	query := orm.GormDB.Model(&orm.ChapterProgress{}).Where("commander_id = ?", commanderID)
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
		_ = ctx.JSON(response.Error("internal_error", "failed to count chapter progress", nil))
		return
	}
	var progress []orm.ChapterProgress
	query = query.Order("updated_at desc")
	query = orm.ApplyPagination(query, meta.Offset, meta.Limit)
	if err := query.Find(&progress).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load chapter progress", nil))
		return
	}
	entries := make([]types.PlayerChapterProgressResponse, 0, len(progress))
	for _, entry := range progress {
		entries = append(entries, types.PlayerChapterProgressResponse{Progress: buildChapterProgressDTO(&entry)})
	}
	_ = ctx.JSON(response.Success(types.PlayerChapterProgressListResponse{Progress: entries, Meta: meta}))
}

// CreatePlayerChapterProgress godoc
// @Summary     Create player chapter progress
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerChapterProgressCreateRequest  true  "Chapter progress payload"
// @Success     200  {object}  PlayerChapterProgressResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-progress [post]
func (handler *PlayerHandler) CreatePlayerChapterProgress(ctx iris.Context) {
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
	var req types.PlayerChapterProgressCreateRequest
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
	if req.Progress.ChapterID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "chapter_id required", nil))
		return
	}
	progress := buildChapterProgressModel(commanderID, req.Progress)
	if err := orm.UpsertChapterProgress(orm.GormDB, progress); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to store chapter progress", nil))
		return
	}
	payload := types.PlayerChapterProgressResponse{Progress: buildChapterProgressDTO(progress)}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerChapterProgress godoc
// @Summary     Update player chapter progress
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       chapter_id  path  int  true  "Chapter ID"
// @Param       payload  body  types.PlayerChapterProgressUpdateRequest  true  "Chapter progress payload"
// @Success     200  {object}  PlayerChapterProgressResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-progress/{chapter_id} [patch]
func (handler *PlayerHandler) UpdatePlayerChapterProgress(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	chapterID, err := parsePathUint32(ctx.Params().Get("chapter_id"), "chapter_id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerChapterProgressUpdateRequest
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
	if req.Progress.ChapterID == 0 {
		req.Progress.ChapterID = chapterID
	}
	progress := buildChapterProgressModel(commanderID, req.Progress)
	if err := orm.UpsertChapterProgress(orm.GormDB, progress); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update chapter progress", nil))
		return
	}
	payload := types.PlayerChapterProgressResponse{Progress: buildChapterProgressDTO(progress)}
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerChapterProgress godoc
// @Summary     Delete player chapter progress
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       chapter_id  path  int  true  "Chapter ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/chapter-progress/{chapter_id} [delete]
func (handler *PlayerHandler) DeletePlayerChapterProgress(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	chapterID, err := parsePathUint32(ctx.Params().Get("chapter_id"), "chapter_id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.First(&orm.Commander{}, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}
	if err := orm.DeleteChapterProgress(orm.GormDB, commanderID, chapterID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete chapter progress", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

func buildChapterProgressDTO(progress *orm.ChapterProgress) types.ChapterProgress {
	return types.ChapterProgress{
		ChapterID:        progress.ChapterID,
		Progress:         progress.Progress,
		KillBossCount:    progress.KillBossCount,
		KillEnemyCount:   progress.KillEnemyCount,
		TakeBoxCount:     progress.TakeBoxCount,
		DefeatCount:      progress.DefeatCount,
		TodayDefeatCount: progress.TodayDefeatCount,
		PassCount:        progress.PassCount,
		UpdatedAt:        progress.UpdatedAt,
	}
}

func buildChapterProgressModel(commanderID uint32, progress types.ChapterProgress) *orm.ChapterProgress {
	return &orm.ChapterProgress{
		CommanderID:      commanderID,
		ChapterID:        progress.ChapterID,
		Progress:         progress.Progress,
		KillBossCount:    progress.KillBossCount,
		KillEnemyCount:   progress.KillEnemyCount,
		TakeBoxCount:     progress.TakeBoxCount,
		DefeatCount:      progress.DefeatCount,
		TodayDefeatCount: progress.TodayDefeatCount,
		PassCount:        progress.PassCount,
	}
}
