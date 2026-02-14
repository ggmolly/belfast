package handlers

import (
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

// PlayerLoveLetterState godoc
// @Summary     Get player love letter state
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerLoveLetterStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/love-letter [get]
func (handler *PlayerHandler) PlayerLoveLetterState(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.CommanderExists(commanderID); err != nil {
		writeCommanderError(ctx, err)
		return
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(commanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load love letter state", nil))
		return
	}
	payload := types.PlayerLoveLetterStateResponse{
		Medals:         state.Medals,
		ManualLetters:  state.ManualLetters,
		ConvertedItems: state.ConvertedItems,
		RewardedIDs:    state.RewardedIDs,
		LetterContents: state.LetterContents,
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerLoveLetterState godoc
// @Summary     Update player love letter state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerLoveLetterStateUpdateRequest  true  "Love letter state update"
// @Success     200  {object}  PlayerLoveLetterStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/love-letter [patch]
func (handler *PlayerHandler) UpdatePlayerLoveLetterState(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.CommanderExists(commanderID); err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerLoveLetterStateUpdateRequest
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
	if req.Medals == nil && req.ManualLetters == nil && req.ConvertedItems == nil && req.RewardedIDs == nil && req.LetterContents == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	state, err := orm.GetOrCreateCommanderLoveLetterState(commanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load love letter state", nil))
		return
	}
	if req.Medals != nil {
		state.Medals = *req.Medals
	}
	if req.ManualLetters != nil {
		state.ManualLetters = *req.ManualLetters
	}
	if req.ConvertedItems != nil {
		state.ConvertedItems = *req.ConvertedItems
	}
	if req.RewardedIDs != nil {
		state.RewardedIDs = *req.RewardedIDs
	}
	if req.LetterContents != nil {
		state.LetterContents = *req.LetterContents
	}
	if err := orm.SaveCommanderLoveLetterState(state); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update love letter state", nil))
		return
	}
	payload := types.PlayerLoveLetterStateResponse{
		Medals:         state.Medals,
		ManualLetters:  state.ManualLetters,
		ConvertedItems: state.ConvertedItems,
		RewardedIDs:    state.RewardedIDs,
		LetterContents: state.LetterContents,
	}
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerLoveLetterState godoc
// @Summary     Delete player love letter state
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/love-letter [delete]
func (handler *PlayerHandler) DeletePlayerLoveLetterState(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid id", nil))
		return
	}
	if err := orm.CommanderExists(commanderID); err != nil {
		writeCommanderError(ctx, err)
		return
	}
	if err := orm.DeleteCommanderLoveLetterState(commanderID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete love letter state", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}
