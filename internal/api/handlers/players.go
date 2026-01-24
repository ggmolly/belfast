package handlers

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/arenashop"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/medalshop"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"github.com/ggmolly/belfast/internal/shopstreet"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

type PlayerHandler struct {
	Validate *validator.Validate
}

func NewPlayerHandler() *PlayerHandler {
	return &PlayerHandler{Validate: validator.New(validator.WithRequiredStructEnabled())}
}

func RegisterPlayerRoutes(party iris.Party, handler *PlayerHandler) {
	party.Get("", handler.ListPlayers)
	party.Get("/search", handler.SearchPlayers)
	party.Get("/{id:uint}", handler.PlayerDetail)
	party.Post("/compensations/push-online", handler.PushOnlineCompensationNotifications)
	party.Get("/{id:uint}/resources", handler.PlayerResources)
	party.Get("/{id:uint}/ships", handler.PlayerShips)
	party.Get("/{id:uint}/items", handler.PlayerItems)
	party.Get("/{id:uint}/builds", handler.PlayerBuilds)
	party.Get("/{id:uint}/builds/queue", handler.PlayerBuildQueue)
	party.Patch("/{id:uint}/builds/counters", handler.UpdatePlayerBuildCounters)
	party.Get("/{id:uint}/mails", handler.PlayerMails)
	party.Get("/{id:uint}/compensations", handler.PlayerCompensations)
	party.Get("/{id:uint}/compensations/{compensation_id:uint}", handler.PlayerCompensation)
	party.Post("/{id:uint}/compensations/push", handler.PushCompensationNotification)
	party.Get("/{id:uint}/tb", handler.PlayerTB)
	party.Post("/{id:uint}/tb", handler.CreatePlayerTB)
	party.Put("/{id:uint}/tb", handler.UpdatePlayerTB)
	party.Delete("/{id:uint}/tb", handler.DeletePlayerTB)
	party.Get("/{id:uint}/fleets", handler.PlayerFleets)
	party.Get("/{id:uint}/skins", handler.PlayerSkins)
	party.Get("/{id:uint}/buffs", handler.PlayerBuffs)
	party.Get("/{id:uint}/flags", handler.PlayerFlags)
	party.Post("/{id:uint}/flags", handler.AddPlayerFlag)
	party.Delete("/{id:uint}/flags/{flag_id:uint}", handler.DeletePlayerFlag)
	party.Get("/{id:uint}/guide", handler.PlayerGuide)
	party.Patch("/{id:uint}/guide", handler.UpdatePlayerGuide)
	party.Get("/{id:uint}/stories", handler.PlayerStories)
	party.Post("/{id:uint}/stories", handler.AddPlayerStory)
	party.Delete("/{id:uint}/stories/{story_id:uint}", handler.DeletePlayerStory)
	party.Get("/{id:uint}/attires", handler.PlayerAttires)
	party.Post("/{id:uint}/attires", handler.AddPlayerAttire)
	party.Delete("/{id:uint}/attires/{type:uint}/{attire_id:uint}", handler.DeletePlayerAttire)
	party.Patch("/{id:uint}/attires/selected", handler.UpdatePlayerAttireSelection)
	party.Get("/{id:uint}/livingarea-covers", handler.PlayerLivingAreaCovers)
	party.Post("/{id:uint}/livingarea-covers", handler.AddPlayerLivingAreaCover)
	party.Delete("/{id:uint}/livingarea-covers/{cover_id:uint}", handler.DeletePlayerLivingAreaCover)
	party.Patch("/{id:uint}/livingarea-covers/selected", handler.UpdatePlayerLivingAreaCover)
	party.Get("/{id:uint}/shopping-street", handler.PlayerShoppingStreet)
	party.Post("/{id:uint}/shopping-street/refresh", handler.RefreshPlayerShoppingStreet)
	party.Put("/{id:uint}/shopping-street", handler.UpdatePlayerShoppingStreet)
	party.Put("/{id:uint}/shopping-street/goods", handler.ReplacePlayerShoppingStreetGoods)
	party.Patch("/{id:uint}/shopping-street/goods/{goods_id:uint}", handler.UpdatePlayerShoppingStreetGood)
	party.Delete("/{id:uint}/shopping-street/goods/{goods_id:uint}", handler.DeletePlayerShoppingStreetGood)
	party.Get("/{id:uint}/arena-shop", handler.PlayerArenaShop)
	party.Post("/{id:uint}/arena-shop/refresh", handler.RefreshPlayerArenaShop)
	party.Put("/{id:uint}/arena-shop", handler.UpdatePlayerArenaShop)
	party.Get("/{id:uint}/medal-shop", handler.PlayerMedalShop)
	party.Post("/{id:uint}/medal-shop/refresh", handler.RefreshPlayerMedalShop)
	party.Put("/{id:uint}/medal-shop", handler.UpdatePlayerMedalShop)
	party.Post("/{id:uint}/ban", handler.BanPlayer)
	party.Delete("/{id:uint}/ban", handler.UnbanPlayer)
	party.Post("/{id:uint}/kick", handler.KickPlayer)
	party.Put("/{id:uint}/resources", handler.UpdateResources)
	party.Post("/{id:uint}/give-ship", handler.GiveShip)
	party.Post("/{id:uint}/give-item", handler.GiveItem)
	party.Post("/{id:uint}/send-mail", handler.SendMail)
	party.Post("/{id:uint}/compensations", handler.CreateCompensation)
	party.Patch("/{id:uint}/compensations/{compensation_id:uint}", handler.UpdateCompensation)
	party.Delete("/{id:uint}/compensations/{compensation_id:uint}", handler.DeleteCompensation)
	party.Post("/{id:uint}/give-skin", handler.GiveSkin)
	party.Post("/{id:uint}/buffs", handler.AddPlayerBuff)
	party.Delete("/{id:uint}/buffs/{buff_id:uint}", handler.DeletePlayerBuff)
	party.Delete("/{id:uint}", handler.DeletePlayer)
}

func parsePagination(ctx iris.Context) (types.PaginationMeta, error) {
	offset, err := parseQueryInt(ctx.URLParamDefault("offset", "0"))
	if err != nil || offset < 0 {
		return types.PaginationMeta{}, fmt.Errorf("offset must be >= 0")
	}
	limit, err := parseQueryInt(ctx.URLParamDefault("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 {
		return types.PaginationMeta{}, fmt.Errorf("limit must be >= 1")
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return types.PaginationMeta{Offset: offset, Limit: limit}, nil
}

// ListPlayers godoc
// @Summary     List players
// @Tags        Players
// @Produce     json
// @Param       offset    query  int     false  "Pagination offset"
// @Param       limit     query  int     false  "Pagination limit"
// @Param       sort      query  string  false  "Sort by last_login"
// @Param       filter    query  string  false  "Filters: online, banned"
// @Param       min_level query  int     false  "Minimum level"
// @Success     200  {object}  ListPlayersResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players [get]
func (handler *PlayerHandler) ListPlayers(ctx iris.Context) {
	params, err := parsePlayerQuery(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result, err := orm.ListCommanders(orm.GormDB, params)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list players", nil))
		return
	}

	payload, err := handler.playerListResponse(result, params)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *PlayerHandler) SearchPlayers(ctx iris.Context) {
	params, err := parsePlayerQuery(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	params.Search = ctx.URLParam("q")
	if strings.TrimSpace(params.Search) == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "search query is required", nil))
		return
	}

	result, err := orm.SearchCommanders(orm.GormDB, params)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to search players", nil))
		return
	}

	payload, err := handler.playerListResponse(result, params)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerDetail godoc
// @Summary     Get player details
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerDetailResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id} [get]
func (handler *PlayerHandler) PlayerDetail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	onlineIDs := onlineCommanderIDs()
	banStatus, err := orm.GetBanStatus(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load ban status", nil))
		return
	}

	payload := types.PlayerDetailResponse{
		CommanderID: commander.CommanderID,
		AccountID:   commander.AccountID,
		Name:        commander.Name,
		Level:       commander.Level,
		Exp:         commander.Exp,
		LastLogin:   commander.LastLogin.UTC().Format(time.RFC3339),
		Banned:      banStatus.Banned,
		Online:      onlineIDs[commander.CommanderID],
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerResources godoc
// @Summary     Get player resources
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerResourcesResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/resources [get]
func (handler *PlayerHandler) PlayerResources(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var allResources []orm.Resource
	if err := orm.GormDB.Find(&allResources).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load resources", nil))
		return
	}

	resourceMap := make(map[uint32]uint32)
	for _, owned := range commander.OwnedResources {
		resourceMap[owned.ResourceID] = owned.Amount
	}

	payload := types.PlayerResourceResponse{Resources: make([]types.PlayerResourceEntry, 0, len(allResources))}
	for _, res := range allResources {
		payload.Resources = append(payload.Resources, types.PlayerResourceEntry{
			ResourceID: res.ID,
			Amount:     resourceMap[res.ID],
			Name:       res.Name,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerItems godoc
// @Summary     Get player items
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerItemsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/items [get]
func (handler *PlayerHandler) PlayerItems(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var allItems []orm.Item
	if err := orm.GormDB.Find(&allItems).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load items", nil))
		return
	}

	itemMap := make(map[uint32]uint32)
	for _, item := range commander.Items {
		itemMap[item.ItemID] = item.Count
	}
	for _, misc := range commander.MiscItems {
		itemMap[misc.ItemID] += misc.Data
	}

	payload := types.PlayerItemResponse{Items: make([]types.PlayerItemEntry, 0, len(allItems))}
	for _, item := range allItems {
		payload.Items = append(payload.Items, types.PlayerItemEntry{
			ItemID: item.ID,
			Count:  itemMap[item.ID],
			Name:   item.Name,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerShips godoc
// @Summary     Get player ships
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerShipsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ships [get]
func (handler *PlayerHandler) PlayerShips(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerShipResponse{Ships: make([]types.PlayerShipEntry, 0, len(commander.Ships))}
	for _, ship := range commander.Ships {
		payload.Ships = append(payload.Ships, types.PlayerShipEntry{
			OwnedID: ship.ID,
			ShipID:  ship.ShipID,
			Level:   ship.Level,
			Rarity:  ship.Ship.RarityID,
			Name:    ship.Ship.Name,
			SkinID:  ship.SkinID,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerBuilds godoc
// @Summary     Get player builds
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerBuildsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds [get]
func (handler *PlayerHandler) PlayerBuilds(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerBuildResponse{Builds: make([]types.PlayerBuildEntry, 0, len(commander.Builds))}
	orderedBuilds := orm.OrderedBuilds(commander.Builds)
	for _, build := range orderedBuilds {
		payload.Builds = append(payload.Builds, types.PlayerBuildEntry{
			BuildID:    build.ID,
			ShipID:     build.ShipID,
			ShipName:   build.Ship.Name,
			PoolID:     build.PoolID,
			FinishesAt: build.FinishesAt.UTC().Format(time.RFC3339),
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerBuildQueue godoc
// @Summary     Get player build queue snapshot
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerBuildQueueResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds/queue [get]
func (handler *PlayerHandler) PlayerBuildQueue(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := buildQueueResponse(commander)
	_ = ctx.JSON(response.Success(payload))
}

func buildQueueResponse(commander orm.Commander) types.PlayerBuildQueueResponse {
	now := time.Now()
	orderedBuilds := orm.OrderedBuilds(commander.Builds)
	queue := make([]types.PlayerBuildQueueEntry, len(orderedBuilds))
	for i, build := range orderedBuilds {
		queue[i] = types.PlayerBuildQueueEntry{
			Slot:             uint32(i + 1),
			PoolID:           build.PoolID,
			RemainingSeconds: orm.RemainingSeconds(build.FinishesAt, now),
			FinishTime:       uint32(build.FinishesAt.Unix()),
		}
	}

	return types.PlayerBuildQueueResponse{
		WorklistCount: consts.MaxBuildWorkCount,
		WorklistList:  queue,
		DrawCount1:    commander.DrawCount1,
		DrawCount10:   commander.DrawCount10,
		ExchangeCount: commander.ExchangeCount,
	}
}

// UpdatePlayerBuildCounters godoc
// @Summary     Update player build counters
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerBuildCounterUpdateRequest  true  "Build counters update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds/counters [patch]
func (handler *PlayerHandler) UpdatePlayerBuildCounters(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.PlayerBuildCounterUpdateRequest
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

	updated := false
	if req.DrawCount1 != nil {
		commander.DrawCount1 = *req.DrawCount1
		updated = true
	}
	if req.DrawCount10 != nil {
		commander.DrawCount10 = *req.DrawCount10
		updated = true
	}
	if req.ExchangeCount != nil {
		commander.ExchangeCount = *req.ExchangeCount
		updated = true
	}
	if !updated {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Save(commander).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update counters", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// PlayerMails godoc
// @Summary     Get player mails
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerMailsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/mails [get]
func (handler *PlayerHandler) PlayerMails(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerMailResponse{Mails: make([]types.PlayerMailEntry, 0, len(commander.Mails))}
	for _, mail := range commander.Mails {
		attachments := make([]types.PlayerMailAttachment, 0, len(mail.Attachments))
		for _, attachment := range mail.Attachments {
			attachments = append(attachments, types.PlayerMailAttachment{
				Type:     attachment.Type,
				ItemID:   attachment.ItemID,
				Quantity: attachment.Quantity,
			})
		}
		payload.Mails = append(payload.Mails, types.PlayerMailEntry{
			MailID:      mail.ID,
			Title:       mail.Title,
			Body:        mail.Body,
			Read:        mail.Read,
			Date:        mail.Date.UTC().Format(time.RFC3339),
			Important:   mail.IsImportant,
			Archived:    mail.IsArchived,
			Sender:      mail.CustomSender,
			Attachments: attachments,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerCompensations godoc
// @Summary     Get player compensations
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerCompensationsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/compensations [get]
func (handler *PlayerHandler) PlayerCompensations(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerCompensationResponse{Compensations: make([]types.PlayerCompensationEntry, 0, len(commander.Compensations))}
	for _, compensation := range commander.Compensations {
		payload.Compensations = append(payload.Compensations, buildCompensationEntry(compensation))
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerCompensation godoc
// @Summary     Get player compensation
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       compensation_id   path  int  true  "Compensation ID"
// @Success     200  {object}  PlayerCompensationsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/compensations/{compensation_id} [get]
func (handler *PlayerHandler) PlayerCompensation(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	compensationID, err := parseCompensationID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var compensation orm.Compensation
	if err := orm.GormDB.Preload("Attachments").Where("commander_id = ?", commanderID).First(&compensation, compensationID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerCompensationResponse{Compensations: []types.PlayerCompensationEntry{buildCompensationEntry(compensation)}}
	_ = ctx.JSON(response.Success(payload))
}

// CreateCompensation godoc
// @Summary     Create compensation for player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.CreateCompensationRequest  true  "Compensation request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/compensations [post]
func (handler *PlayerHandler) CreateCompensation(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.CreateCompensationRequest
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

	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "expires_at must be RFC3339", nil))
		return
	}
	sendTime := time.Now()
	if strings.TrimSpace(req.SendTime) != "" {
		parsed, err := time.Parse(time.RFC3339, req.SendTime)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "send_time must be RFC3339", nil))
			return
		}
		sendTime = parsed
	}

	compensation := orm.Compensation{
		CommanderID: commander.CommanderID,
		Title:       req.Title,
		Text:        req.Text,
		SendTime:    sendTime,
		ExpiresAt:   expiresAt,
	}
	for _, attachment := range req.Attachments {
		compensation.Attachments = append(compensation.Attachments, orm.CompensationAttachment{
			Type:     attachment.Type,
			ItemID:   attachment.ItemID,
			Quantity: attachment.Quantity,
		})
	}

	if err := orm.GormDB.Create(&compensation).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create compensation", nil))
		return
	}

	if client := findCommanderClient(commander.CommanderID); client != nil {
		if err := sendCompensationNotification(client); err != nil {
			logger.LogEvent("API", "CompensationPush", fmt.Sprintf("failed to push notification: %v", err), logger.LOG_LEVEL_ERROR)
		}
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateCompensation godoc
// @Summary     Update compensation for player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       compensation_id   path  int  true  "Compensation ID"
// @Param       payload  body  types.UpdateCompensationRequest  true  "Compensation update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/compensations/{compensation_id} [patch]
func (handler *PlayerHandler) UpdateCompensation(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	compensationID, err := parseCompensationID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.UpdateCompensationRequest
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

	var compensation orm.Compensation
	if err := orm.GormDB.Preload("Attachments").Where("commander_id = ?", commanderID).First(&compensation, compensationID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}

	if req.Title != nil {
		compensation.Title = *req.Title
	}
	if req.Text != nil {
		compensation.Text = *req.Text
	}
	if req.SendTime != nil {
		parsed, err := time.Parse(time.RFC3339, *req.SendTime)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "send_time must be RFC3339", nil))
			return
		}
		compensation.SendTime = parsed
	}
	if req.ExpiresAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "expires_at must be RFC3339", nil))
			return
		}
		compensation.ExpiresAt = parsed
	}
	if req.AttachFlag != nil {
		compensation.AttachFlag = *req.AttachFlag
	}
	if req.Attachments != nil {
		if err := orm.GormDB.Where("compensation_id = ?", compensation.ID).Delete(&orm.CompensationAttachment{}).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update attachments", nil))
			return
		}
		compensation.Attachments = make([]orm.CompensationAttachment, 0, len(*req.Attachments))
		for _, attachment := range *req.Attachments {
			compensation.Attachments = append(compensation.Attachments, orm.CompensationAttachment{
				Type:     attachment.Type,
				ItemID:   attachment.ItemID,
				Quantity: attachment.Quantity,
			})
		}
	}

	if err := orm.GormDB.Session(&gorm.Session{FullSaveAssociations: true}).Save(&compensation).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update compensation", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteCompensation godoc
// @Summary     Delete compensation from player
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       compensation_id   path  int  true  "Compensation ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/compensations/{compensation_id} [delete]
func (handler *PlayerHandler) DeleteCompensation(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	compensationID, err := parseCompensationID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if err := orm.GormDB.Where("commander_id = ? AND id = ?", commanderID, compensationID).Delete(&orm.Compensation{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete compensation", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// PushCompensationNotification godoc
// @Summary     Push compensation notification
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/compensations/push [post]
func (handler *PlayerHandler) PushCompensationNotification(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	client := findCommanderClient(commanderID)
	if client == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "player not online", nil))
		return
	}

	if err := sendCompensationNotification(client); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to push notification", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// PushOnlineCompensationNotifications godoc
// @Summary     Push compensation notifications to online players
// @Tags        Players
// @Produce     json
// @Success     200  {object}  PushCompensationResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/compensations/push-online [post]
func (handler *PlayerHandler) PushOnlineCompensationNotifications(ctx iris.Context) {
	clients := connection.BelfastInstance.ListClients()
	if len(clients) == 0 {
		_ = ctx.JSON(response.Success(types.PushCompensationResponse{Pushed: 0, Failed: 0}))
		return
	}

	var pushed int
	var failed int
	for _, client := range clients {
		if client.Commander == nil {
			continue
		}
		if err := sendCompensationNotification(client); err != nil {
			failed++
			logger.LogEvent("API", "CompensationPush", fmt.Sprintf("failed to push notification: %v", err), logger.LOG_LEVEL_ERROR)
			continue
		}
		pushed++
	}

	_ = ctx.JSON(response.Success(types.PushCompensationResponse{Pushed: pushed, Failed: failed}))
}

// PlayerFleets godoc
// @Summary     Get player fleets
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerFleetsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/fleets [get]
func (handler *PlayerHandler) PlayerFleets(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerFleetResponse{Fleets: make([]types.PlayerFleetEntry, 0, len(commander.Fleets))}
	for _, fleet := range commander.Fleets {
		ships := make([]uint32, 0, len(fleet.ShipList))
		for _, shipID := range fleet.ShipList {
			ships = append(ships, uint32(shipID))
		}
		payload.Fleets = append(payload.Fleets, types.PlayerFleetEntry{
			FleetID: fleet.GameID,
			Name:    fleet.Name,
			Ships:   ships,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerSkins godoc
// @Summary     Get player skins
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerSkinsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/skins [get]
func (handler *PlayerHandler) PlayerSkins(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var skins []orm.Skin
	if err := orm.GormDB.Find(&skins).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load skins", nil))
		return
	}

	skinOwned := make(map[uint32]*time.Time, len(commander.OwnedSkins))
	for _, owned := range commander.OwnedSkins {
		skinOwned[owned.SkinID] = owned.ExpiresAt
	}

	payload := types.PlayerSkinResponse{Skins: make([]types.PlayerSkinEntry, 0, len(skins))}
	for _, skin := range skins {
		expiresAt := skinOwned[skin.ID]
		payload.Skins = append(payload.Skins, types.PlayerSkinEntry{
			SkinID:    skin.ID,
			Name:      skin.Name,
			ExpiresAt: expiresAt,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerBuffs godoc
// @Summary     Get player buffs
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       active  query  bool  false  "Only active buffs"
// @Success     200  {object}  PlayerBuffsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/buffs [get]
func (handler *PlayerHandler) PlayerBuffs(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	activeOnly, err := parseOptionalBool(ctx.URLParam("active"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var buffs []orm.CommanderBuff
	if activeOnly {
		buffs, err = orm.ListCommanderActiveBuffs(commander.CommanderID, time.Now().UTC())
	} else {
		buffs, err = orm.ListCommanderBuffs(commander.CommanderID)
	}
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load buffs", nil))
		return
	}

	payload := types.PlayerBuffResponse{Buffs: make([]types.PlayerBuffEntry, 0, len(buffs))}
	for _, buff := range buffs {
		payload.Buffs = append(payload.Buffs, types.PlayerBuffEntry{
			BuffID:    buff.BuffID,
			ExpiresAt: buff.ExpiresAt.UTC().Format(time.RFC3339),
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerFlags godoc
// @Summary     Get player common flags
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerFlagsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/flags [get]
func (handler *PlayerHandler) PlayerFlags(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	flags, err := orm.ListCommanderCommonFlags(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load flags", nil))
		return
	}
	payload := types.PlayerFlagsResponse{Flags: flags}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerFlag godoc
// @Summary     Add player common flag
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerFlagRequest  true  "Flag request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/flags [post]
func (handler *PlayerHandler) AddPlayerFlag(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerFlagRequest
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
	if err := orm.SetCommanderCommonFlag(orm.GormDB, commander.CommanderID, req.FlagID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to add flag", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerFlag godoc
// @Summary     Remove player common flag
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       flag_id   path  int  true  "Flag ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/flags/{flag_id} [delete]
func (handler *PlayerHandler) DeletePlayerFlag(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	flagID, err := parsePathUint32(ctx.Params().Get("flag_id"), "flag id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.ClearCommanderCommonFlag(orm.GormDB, commander.CommanderID, flagID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete flag", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerGuide godoc
// @Summary     Get player guide indices
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerGuideResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guide [get]
func (handler *PlayerHandler) PlayerGuide(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	payload := types.PlayerGuideResponse{
		GuideIndex:    commander.GuideIndex,
		NewGuideIndex: commander.NewGuideIndex,
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerGuide godoc
// @Summary     Update player guide indices
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerGuideUpdateRequest  true  "Guide update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guide [patch]
func (handler *PlayerHandler) UpdatePlayerGuide(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerGuideUpdateRequest
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
	updates := map[string]interface{}{}
	if req.GuideIndex != nil {
		commander.GuideIndex = *req.GuideIndex
		updates["guide_index"] = *req.GuideIndex
	}
	if req.NewGuideIndex != nil {
		commander.NewGuideIndex = *req.NewGuideIndex
		updates["new_guide_index"] = *req.NewGuideIndex
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(commander).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update guide indices", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerStories godoc
// @Summary     Get player story progress
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerStoriesResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/stories [get]
func (handler *PlayerHandler) PlayerStories(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	stories, err := orm.ListCommanderStoryIDs(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load stories", nil))
		return
	}
	payload := types.PlayerStoriesResponse{Stories: stories}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerStory godoc
// @Summary     Add player story progress
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerStoryRequest  true  "Story request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/stories [post]
func (handler *PlayerHandler) AddPlayerStory(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerStoryRequest
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
	if err := orm.AddCommanderStory(orm.GormDB, commander.CommanderID, req.StoryID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to add story", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerStory godoc
// @Summary     Remove player story progress
// @Tags        Players
// @Produce     json
// @Param       id         path  int  true  "Player ID"
// @Param       story_id   path  int  true  "Story ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/stories/{story_id} [delete]
func (handler *PlayerHandler) DeletePlayerStory(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	storyID, err := parsePathUint32(ctx.Params().Get("story_id"), "story id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.Where("commander_id = ? AND story_id = ?", commander.CommanderID, storyID).Delete(&orm.CommanderStory{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete story", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerAttires godoc
// @Summary     Get player attires
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerAttiresResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/attires [get]
func (handler *PlayerHandler) PlayerAttires(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	entries, err := orm.ListCommanderAttires(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load attires", nil))
		return
	}
	payload := types.PlayerAttireResponse{Attires: make([]types.PlayerAttireEntry, 0, len(entries))}
	for _, entry := range entries {
		payload.Attires = append(payload.Attires, types.PlayerAttireEntry{
			Type:      entry.Type,
			AttireID:  entry.AttireID,
			ExpiresAt: entry.ExpiresAt,
			IsNew:     entry.IsNew,
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerAttire godoc
// @Summary     Add player attire
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerAttireCreateRequest  true  "Attire request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/attires [post]
func (handler *PlayerHandler) AddPlayerAttire(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerAttireCreateRequest
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
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "expires_at must be RFC3339", nil))
			return
		}
		expiresAt = &parsed
	}
	entry := orm.CommanderAttire{
		CommanderID: commander.CommanderID,
		Type:        req.Type,
		AttireID:    req.AttireID,
		ExpiresAt:   expiresAt,
	}
	if req.IsNew != nil {
		entry.IsNew = *req.IsNew
	}
	if err := orm.UpsertCommanderAttire(orm.GormDB, entry); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to add attire", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerAttire godoc
// @Summary     Remove player attire
// @Tags        Players
// @Produce     json
// @Param       id          path  int  true  "Player ID"
// @Param       type        path  int  true  "Attire type"
// @Param       attire_id   path  int  true  "Attire ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/attires/{type}/{attire_id} [delete]
func (handler *PlayerHandler) DeletePlayerAttire(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	attireType, err := parsePathUint32(ctx.Params().Get("type"), "attire type")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	attireID, err := parsePathUint32(ctx.Params().Get("attire_id"), "attire id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.Where("commander_id = ? AND type = ? AND attire_id = ?", commander.CommanderID, attireType, attireID).Delete(&orm.CommanderAttire{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete attire", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdatePlayerAttireSelection godoc
// @Summary     Update player attire selection
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerAttireSelectionUpdateRequest  true  "Selection update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/attires/selected [patch]
func (handler *PlayerHandler) UpdatePlayerAttireSelection(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerAttireSelectionUpdateRequest
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
	updates := map[string]interface{}{}
	if req.IconFrameID != nil {
		commander.SelectedIconFrameID = *req.IconFrameID
		updates["selected_icon_frame_id"] = *req.IconFrameID
	}
	if req.ChatFrameID != nil {
		commander.SelectedChatFrameID = *req.ChatFrameID
		updates["selected_chat_frame_id"] = *req.ChatFrameID
	}
	if req.BattleUIID != nil {
		commander.SelectedBattleUIID = *req.BattleUIID
		updates["selected_battle_ui_id"] = *req.BattleUIID
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(commander).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update attire selection", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerLivingAreaCovers godoc
// @Summary     Get player living area covers
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerLivingAreaCoverResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/livingarea-covers [get]
func (handler *PlayerHandler) PlayerLivingAreaCovers(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	entries, err := orm.ListCommanderLivingAreaCovers(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load covers", nil))
		return
	}
	owned := make([]uint32, 0, len(entries))
	for _, entry := range entries {
		owned = append(owned, entry.CoverID)
	}
	if commander.LivingAreaCoverID != 0 {
		found := false
		for _, coverID := range owned {
			if coverID == commander.LivingAreaCoverID {
				found = true
				break
			}
		}
		if !found {
			owned = append(owned, commander.LivingAreaCoverID)
		}
	}
	payload := types.PlayerLivingAreaCoverResponse{
		Selected: commander.LivingAreaCoverID,
		Owned:    owned,
	}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerLivingAreaCover godoc
// @Summary     Add player living area cover
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerLivingAreaCoverRequest  true  "Cover request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/livingarea-covers [post]
func (handler *PlayerHandler) AddPlayerLivingAreaCover(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerLivingAreaCoverRequest
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
	entry := orm.CommanderLivingAreaCover{CommanderID: commander.CommanderID, CoverID: req.CoverID}
	if err := orm.UpsertCommanderLivingAreaCover(orm.GormDB, entry); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to add cover", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerLivingAreaCover godoc
// @Summary     Remove player living area cover
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       cover_id  path  int  true  "Cover ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/livingarea-covers/{cover_id} [delete]
func (handler *PlayerHandler) DeletePlayerLivingAreaCover(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	coverID, err := parsePathUint32(ctx.Params().Get("cover_id"), "cover id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.Where("commander_id = ? AND cover_id = ?", commander.CommanderID, coverID).Delete(&orm.CommanderLivingAreaCover{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete cover", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdatePlayerLivingAreaCover godoc
// @Summary     Update player living area cover selection
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerLivingAreaCoverSelectRequest  true  "Cover selection"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/livingarea-covers/selected [patch]
func (handler *PlayerHandler) UpdatePlayerLivingAreaCover(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerLivingAreaCoverSelectRequest
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
	commander.LivingAreaCoverID = req.CoverID
	if err := orm.GormDB.Model(commander).Update("living_area_cover_id", req.CoverID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update cover selection", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerShoppingStreet godoc
// @Summary     Get player shopping street
// @Tags        Players
// @Produce     json
// @Param       id              path   int   true   "Player ID"
// @Param       include_offers  query  bool  false  "Include offer metadata"
// @Success     200  {object}  PlayerShoppingStreetResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street [get]
func (handler *PlayerHandler) PlayerShoppingStreet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	includeOffers, err := parseOptionalBool(ctx.URLParam("include_offers"))
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	state, goods, err := shopstreet.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street", nil))
		return
	}
	payload, err := buildShoppingStreetResponse(state, goods, includeOffers)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}
	_ = ctx.JSON(response.Success(payload))
}

// RefreshPlayerShoppingStreet godoc
// @Summary     Refresh player shopping street
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       body  body  types.ShoppingStreetRefreshRequest  false  "Refresh settings"
// @Success     200  {object}  PlayerShoppingStreetResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street/refresh [post]
func (handler *PlayerHandler) RefreshPlayerShoppingStreet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.ShoppingStreetRefreshRequest
	if err := ctx.ReadJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if req.GoodsCount != nil && *req.GoodsCount <= 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "goods_count must be >= 1", nil))
		return
	}
	if req.DiscountOverride != nil && (*req.DiscountOverride < 1 || *req.DiscountOverride > 100) {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "discount_override must be between 1 and 100", nil))
		return
	}
	if len(req.GoodsIDs) > 0 {
		_, invalid, err := shopstreet.ResolveOffers(req.GoodsIDs)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to validate goods ids", nil))
			return
		}
		if len(invalid) > 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid goods ids", map[string][]uint32{"invalid_ids": invalid}))
			return
		}
	}
	options := shopstreet.RefreshOptions{
		GoodsCount:         req.GoodsCount,
		NextFlashInSeconds: req.NextFlashInSeconds,
		SetFlashCount:      req.SetFlashCount,
		Seed:               req.Seed,
		GoodsIDs:           req.GoodsIDs,
		DiscountOverride:   req.DiscountOverride,
		BuyCount:           req.BuyCount,
	}
	state, goods, err := shopstreet.RefreshGoods(commander.CommanderID, time.Now(), options)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to refresh shopping street", nil))
		return
	}
	payload, err := buildShoppingStreetResponse(state, goods, false)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerShoppingStreet godoc
// @Summary     Update player shopping street state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       body  body  types.ShoppingStreetUpdateRequest  true  "State updates"
// @Success     200  {object}  PlayerShoppingStreetResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street [put]
func (handler *PlayerHandler) UpdatePlayerShoppingStreet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.ShoppingStreetUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	state, goods, err := shopstreet.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if req.NextFlashTime != nil {
		updates["next_flash_time"] = *req.NextFlashTime
	}
	if req.LevelUpTime != nil {
		updates["level_up_time"] = *req.LevelUpTime
	}
	if req.FlashCount != nil {
		updates["flash_count"] = *req.FlashCount
	}
	if len(updates) > 0 {
		if err := orm.GormDB.Model(&orm.ShoppingStreetState{}).Where("commander_id = ?", commander.CommanderID).Updates(updates).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update shopping street", nil))
			return
		}
		if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to reload shopping street", nil))
			return
		}
	}
	payload, err := buildShoppingStreetResponse(state, goods, false)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}
	_ = ctx.JSON(response.Success(payload))
}

// ReplacePlayerShoppingStreetGoods godoc
// @Summary     Replace player shopping street goods
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       body  body  types.ShoppingStreetGoodsReplaceRequest  true  "Goods list"
// @Success     200  {object}  PlayerShoppingStreetResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street/goods [put]
func (handler *PlayerHandler) ReplacePlayerShoppingStreetGoods(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.ShoppingStreetGoodsReplaceRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	goods := make([]orm.ShoppingStreetGood, 0, len(req.Goods))
	ids := make([]uint32, 0, len(req.Goods))
	for _, item := range req.Goods {
		if item.Discount < 1 || item.Discount > 100 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "discount must be between 1 and 100", nil))
			return
		}
		goods = append(goods, orm.ShoppingStreetGood{
			CommanderID: commander.CommanderID,
			GoodsID:     item.GoodsID,
			Discount:    item.Discount,
			BuyCount:    item.BuyCount,
		})
		ids = append(ids, item.GoodsID)
	}
	if len(ids) > 0 {
		_, invalid, err := shopstreet.ResolveOffers(ids)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to validate goods ids", nil))
			return
		}
		if len(invalid) > 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid goods ids", map[string][]uint32{"invalid_ids": invalid}))
			return
		}
	}
	state, _, err := shopstreet.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street", nil))
		return
	}
	updatedGoods, err := shopstreet.ReplaceGoods(commander.CommanderID, goods)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to replace goods", nil))
		return
	}
	payload, err := buildShoppingStreetResponse(state, updatedGoods, false)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerShoppingStreetGood godoc
// @Summary     Update player shopping street good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       goods_id  path  int  true  "Goods ID"
// @Param       body      body  types.ShoppingStreetGoodPatchRequest  true  "Good updates"
// @Success     200  {object}  PlayerShoppingStreetResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street/goods/{goods_id} [patch]
func (handler *PlayerHandler) UpdatePlayerShoppingStreetGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	goodsID, err := parsePathUint32(ctx.Params().Get("goods_id"), "goods id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	_, invalid, err := shopstreet.ResolveOffers([]uint32{goodsID})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to validate goods id", nil))
		return
	}
	if len(invalid) > 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid goods id", nil))
		return
	}
	var req types.ShoppingStreetGoodPatchRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.Discount != nil {
		if *req.Discount < 1 || *req.Discount > 100 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "discount must be between 1 and 100", nil))
			return
		}
		updates["discount"] = *req.Discount
	}
	if req.BuyCount != nil {
		updates["buy_count"] = *req.BuyCount
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	result := orm.GormDB.Model(&orm.ShoppingStreetGood{}).
		Where("commander_id = ? AND goods_id = ?", commander.CommanderID, goodsID).
		Updates(updates)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update goods", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "goods not found", nil))
		return
	}
	state, goods, err := shopstreet.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street", nil))
		return
	}
	payload, err := buildShoppingStreetResponse(state, goods, false)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerShoppingStreetGood godoc
// @Summary     Delete player shopping street good
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       goods_id  path  int  true  "Goods ID"
// @Success     200  {object}  PlayerShoppingStreetResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street/goods/{goods_id} [delete]
func (handler *PlayerHandler) DeletePlayerShoppingStreetGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	goodsID, err := parsePathUint32(ctx.Params().Get("goods_id"), "goods id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	_, invalid, err := shopstreet.ResolveOffers([]uint32{goodsID})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to validate goods id", nil))
		return
	}
	if len(invalid) > 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid goods id", nil))
		return
	}
	result := orm.GormDB.Where("commander_id = ? AND goods_id = ?", commander.CommanderID, goodsID).Delete(&orm.ShoppingStreetGood{})
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete goods", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "goods not found", nil))
		return
	}
	state, goods, err := shopstreet.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street", nil))
		return
	}
	payload, err := buildShoppingStreetResponse(state, goods, false)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build response", nil))
		return
	}
	_ = ctx.JSON(response.Success(payload))
}

func (handler *PlayerHandler) BanPlayer(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.BanPlayerRequest
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
	if req.Permanent {
		if req.LiftTimestamp != "" || req.DurationSec != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "permanent cannot be combined with lift_timestamp or duration_sec", nil))
			return
		}
	}
	if req.LiftTimestamp != "" && req.DurationSec != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "lift_timestamp and duration_sec cannot both be set", nil))
		return
	}
	if !req.Permanent && req.LiftTimestamp == "" && req.DurationSec == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "ban requires duration_sec, lift_timestamp, or permanent=true", nil))
		return
	}

	if _, err := orm.LoadCommanderWithDetails(commanderID); err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var liftTimestamp *time.Time
	if req.LiftTimestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.LiftTimestamp)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "lift_timestamp must be RFC3339", nil))
			return
		}
		liftTimestamp = &parsed
	}
	if req.DurationSec != nil {
		parsed := time.Now().UTC().Add(time.Duration(*req.DurationSec) * time.Second)
		liftTimestamp = &parsed
	}

	punishment := orm.Punishment{
		PunishedID:    commanderID,
		IsPermanent:   req.Permanent,
		LiftTimestamp: liftTimestamp,
	}
	if err := punishment.Create(); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to ban player", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

func (handler *PlayerHandler) UnbanPlayer(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if _, err := orm.LoadCommanderWithDetails(commanderID); err != nil {
		writeCommanderError(ctx, err)
		return
	}

	punishment, err := orm.ActivePunishment(commanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "player is not banned", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load punishment", nil))
		return
	}
	if err := orm.GormDB.Delete(punishment).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to unban player", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

func (handler *PlayerHandler) KickPlayer(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	client := findCommanderClient(commanderID)
	if client == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "player not online", nil))
		return
	}

	reason := uint8(consts.DR_CONNECTION_TO_SERVER_LOST)
	if ctx.GetContentLength() > 0 {
		var req types.KickPlayerRequest
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
		if req.Reason != 0 {
			reason = req.Reason
		}
	}

	if err := client.Disconnect(reason); err != nil {
		logger.LogEvent("API", "Kick", fmt.Sprintf("failed to send disconnect: %v", err), logger.LOG_LEVEL_ERROR)
	}
	if err := client.Flush(); err != nil {
		logger.LogEvent("API", "Kick", fmt.Sprintf("failed to flush disconnect: %v", err), logger.LOG_LEVEL_ERROR)
	}
	connection.BelfastInstance.RemoveClient(client)

	payload := types.KickPlayerResponse{Disconnected: true}
	_ = ctx.JSON(response.Success(payload))
}

// UpdateResources godoc
// @Summary     Update player resources
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.ResourceUpdateRequest  true  "Resource update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/resources [put]
func (handler *PlayerHandler) UpdateResources(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.ResourceUpdateRequest
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
	if !dedupeResourceUpdates(req.Resources) {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "duplicate resource_id values", nil))
		return
	}

	for _, entry := range req.Resources {
		if err := commander.SetResource(entry.ResourceID, entry.Amount); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update resources", nil))
			return
		}
	}

	_ = ctx.JSON(response.Success(nil))
}

// GiveShip godoc
// @Summary     Give ship to player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.GiveShipRequest  true  "Ship grant"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/give-ship [post]
func (handler *PlayerHandler) GiveShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.GiveShipRequest
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

	if _, err := commander.AddShip(req.ShipID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to give ship", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// GiveItem godoc
// @Summary     Give item to player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.GiveItemRequest  true  "Item grant"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/give-item [post]
func (handler *PlayerHandler) GiveItem(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.GiveItemRequest
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

	if err := commander.AddItem(req.ItemID, req.Amount); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to give item", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// GiveSkin godoc
// @Summary     Give skin to player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.GiveSkinRequest  true  "Skin grant"
// @Success     204  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/give-skin [post]
func (handler *PlayerHandler) GiveSkin(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.GiveSkinRequest
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

	if err := commander.GiveSkin(req.SkinID); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to give skin", nil))
		return
	}

	ctx.StatusCode(iris.StatusNoContent)
}

// AddPlayerBuff godoc
// @Summary     Add buff to player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerBuffAddRequest  true  "Buff grant"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/buffs [post]
func (handler *PlayerHandler) AddPlayerBuff(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.PlayerBuffAddRequest
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

	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "expires_at must be RFC3339", nil))
		return
	}

	if err := orm.UpsertCommanderBuff(commander.CommanderID, req.BuffID, expiresAt); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to add buff", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerBuff godoc
// @Summary     Delete buff from player
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       buff_id  path  int  true  "Buff ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/buffs/{buff_id} [delete]
func (handler *PlayerHandler) DeletePlayerBuff(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	buffIDParam := ctx.Params().Get("buff_id")
	buffID, err := strconv.ParseUint(buffIDParam, 10, 32)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid buff_id", nil))
		return
	}

	if err := orm.GormDB.Where("commander_id = ? AND buff_id = ?", commander.CommanderID, uint32(buffID)).
		Delete(&orm.CommanderBuff{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete buff", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// SendMail godoc
// @Summary     Send mail to player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.SendMailRequest  true  "Mail request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/send-mail [post]
func (handler *PlayerHandler) SendMail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.SendMailRequest
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

	mail := orm.Mail{
		ReceiverID:   commander.CommanderID,
		Title:        req.Title,
		Body:         req.Body,
		CustomSender: req.CustomSender,
	}
	for _, attachment := range req.Attachments {
		mail.Attachments = append(mail.Attachments, orm.MailAttachment{
			Type:     attachment.Type,
			ItemID:   attachment.ItemID,
			Quantity: attachment.Quantity,
		})
	}

	if err := commander.SendMail(&mail); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to send mail", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayer godoc
// @Summary     Delete player
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id} [delete]
func (handler *PlayerHandler) DeletePlayer(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var commander orm.Commander
	if err := orm.GormDB.First(&commander, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}

	if err := orm.GormDB.Delete(&commander).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete player", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

func parsePlayerQuery(ctx iris.Context) (orm.PlayerQueryParams, error) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		return orm.PlayerQueryParams{}, err
	}

	sort := strings.TrimSpace(ctx.URLParam("sort"))
	if sort != "" && sort != "last_login" {
		return orm.PlayerQueryParams{}, fmt.Errorf("unsupported sort")
	}

	filters := strings.TrimSpace(ctx.URLParam("filter"))
	params := orm.PlayerQueryParams{
		Offset:   pagination.Offset,
		Limit:    pagination.Limit,
		MinLevel: parseQueryMinLevel(ctx.URLParam("min_level")),
	}

	switch filters {
	case "":
	case "online":
		params.FilterOnline = true
		params.OnlineIDs = onlineIDSlice()
	case "banned":
		params.FilterBanned = true
	case "online,banned":
		params.FilterOnline = true
		params.FilterBanned = true
		params.OnlineIDs = onlineIDSlice()
	case "banned,online":
		params.FilterOnline = true
		params.FilterBanned = true
		params.OnlineIDs = onlineIDSlice()
	default:
		return orm.PlayerQueryParams{}, fmt.Errorf("unsupported filter")
	}

	return params, nil
}

func parseQueryMinLevel(value string) int {
	if strings.TrimSpace(value) == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}

func parseQueryInt(value string) (int, error) {
	return strconv.Atoi(value)
}

func parseOptionalBool(value string) (bool, error) {
	if strings.TrimSpace(value) == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean value")
	}
	return parsed, nil
}

func loadCommanderDetail(ctx iris.Context) (orm.Commander, error) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		return orm.Commander{}, err
	}
	commander, err := orm.LoadCommanderWithDetails(commanderID)
	if err != nil {
		return orm.Commander{}, err
	}
	commander.OwnedResourcesMap = make(map[uint32]*orm.OwnedResource)
	for i := range commander.OwnedResources {
		resource := &commander.OwnedResources[i]
		commander.OwnedResourcesMap[resource.ResourceID] = resource
	}
	commander.CommanderItemsMap = make(map[uint32]*orm.CommanderItem)
	for i := range commander.Items {
		item := &commander.Items[i]
		commander.CommanderItemsMap[item.ItemID] = item
	}
	commander.MiscItemsMap = make(map[uint32]*orm.CommanderMiscItem)
	for i := range commander.MiscItems {
		item := &commander.MiscItems[i]
		commander.MiscItemsMap[item.ItemID] = item
	}
	commander.OwnedSkinsMap = make(map[uint32]*orm.OwnedSkin)
	for i := range commander.OwnedSkins {
		skin := &commander.OwnedSkins[i]
		commander.OwnedSkinsMap[skin.SkinID] = skin
	}
	commander.OwnedShipsMap = make(map[uint32]*orm.OwnedShip)
	for i := range commander.Ships {
		ship := &commander.Ships[i]
		commander.OwnedShipsMap[ship.ID] = ship
	}
	return commander, nil
}

func parseCommanderID(ctx iris.Context) (uint32, error) {
	idParam := ctx.Params().Get("id")
	parsed, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid id")
	}
	return uint32(parsed), nil
}

func parseCompensationID(ctx iris.Context) (uint32, error) {
	value := ctx.Params().Get("compensation_id")
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid compensation_id")
	}
	return uint32(parsed), nil
}

func buildCompensationEntry(compensation orm.Compensation) types.PlayerCompensationEntry {
	attachments := make([]types.PlayerCompensationAttachment, 0, len(compensation.Attachments))
	for _, attachment := range compensation.Attachments {
		attachments = append(attachments, types.PlayerCompensationAttachment{
			Type:     attachment.Type,
			ItemID:   attachment.ItemID,
			Quantity: attachment.Quantity,
		})
	}

	sendTime := ""
	if !compensation.SendTime.IsZero() {
		sendTime = compensation.SendTime.UTC().Format(time.RFC3339)
	}
	expiresAt := ""
	if !compensation.ExpiresAt.IsZero() {
		expiresAt = compensation.ExpiresAt.UTC().Format(time.RFC3339)
	}

	return types.PlayerCompensationEntry{
		CompensationID: compensation.ID,
		Title:          compensation.Title,
		Text:           compensation.Text,
		SendTime:       sendTime,
		ExpiresAt:      expiresAt,
		AttachFlag:     compensation.AttachFlag,
		Attachments:    attachments,
	}
}

func sendCompensationNotification(client *connection.Client) error {
	compensations, err := orm.LoadCommanderCompensations(client.Commander.CommanderID)
	if err != nil {
		return err
	}
	client.Commander.Compensations = compensations
	client.Commander.CompensationsMap = make(map[uint32]*orm.Compensation)
	for i := range client.Commander.Compensations {
		compensation := &client.Commander.Compensations[i]
		client.Commander.CompensationsMap[compensation.ID] = compensation
	}
	buffer := []byte{}
	_, _, err = answer.CompensateNotification(&buffer, client)
	return err
}

func writeCommanderError(ctx iris.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "player not found", nil))
		return
	}
	ctx.StatusCode(iris.StatusInternalServerError)
	_ = ctx.JSON(response.Error("internal_error", "failed to load player", nil))
}

func validationErrors(err error) []map[string]string {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}
	issues := make([]map[string]string, 0, len(validationErrs))
	for _, v := range validationErrs {
		issues = append(issues, map[string]string{
			"field": v.Field(),
			"tag":   v.Tag(),
		})
	}
	return issues
}

// PlayerArenaShop godoc
// @Summary     Get player arena shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerArenaShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/arena-shop [get]
func (handler *PlayerHandler) PlayerArenaShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	config, err := arenashop.LoadConfig()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load arena shop config", nil))
		return
	}
	state, err := arenashop.RefreshIfNeeded(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load arena shop state", nil))
		return
	}
	shopList := arenashop.BuildShopList(state.FlashCount, config)
	payload := buildArenaShopResponse(state, shopList)
	_ = ctx.JSON(response.Success(payload))
}

// RefreshPlayerArenaShop godoc
// @Summary     Refresh player arena shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerArenaShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/arena-shop/refresh [post]
func (handler *PlayerHandler) RefreshPlayerArenaShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	config, err := arenashop.LoadConfig()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load arena shop config", nil))
		return
	}
	state, err := arenashop.RefreshIfNeeded(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load arena shop state", nil))
		return
	}
	refreshCount := int(state.FlashCount + 1)
	if refreshCount > len(config.Template.RefreshPrice) {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "refresh limit reached", nil))
		return
	}
	refreshCost := config.Template.RefreshPrice[refreshCount-1]
	if refreshCost > 0 && !commander.HasEnoughResource(4, refreshCost) {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "insufficient gems", nil))
		return
	}
	updatedState, shopList, cost, err := arenashop.RefreshShop(commander.CommanderID, time.Now(), config)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to refresh arena shop", nil))
		return
	}
	if cost > 0 {
		if err := commander.ConsumeResource(4, cost); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to consume gems", nil))
			return
		}
	}
	payload := buildArenaShopResponse(updatedState, shopList)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerArenaShop godoc
// @Summary     Update player arena shop state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id    path  int  true  "Player ID"
// @Param       body  body  types.ArenaShopUpdateRequest  true  "State updates"
// @Success     200  {object}  PlayerArenaShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/arena-shop [put]
func (handler *PlayerHandler) UpdatePlayerArenaShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.ArenaShopUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	state, err := arenashop.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load arena shop state", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.FlashCount != nil {
		updates["flash_count"] = *req.FlashCount
	}
	if req.NextFlashTime != nil {
		updates["next_flash_time"] = *req.NextFlashTime
	}
	if req.LastRefreshTime != nil {
		updates["last_refresh_time"] = *req.LastRefreshTime
	}
	if len(updates) > 0 {
		if err := orm.GormDB.Model(&orm.ArenaShopState{}).
			Where("commander_id = ?", commander.CommanderID).
			Updates(updates).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update arena shop state", nil))
			return
		}
		if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to reload arena shop state", nil))
			return
		}
	}
	config, err := arenashop.LoadConfig()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load arena shop config", nil))
		return
	}
	shopList := arenashop.BuildShopList(state.FlashCount, config)
	payload := buildArenaShopResponse(state, shopList)
	_ = ctx.JSON(response.Success(payload))
}

func buildArenaShopResponse(state *orm.ArenaShopState, items []*protobuf.ARENASHOP) types.ArenaShopResponse {
	response := types.ArenaShopResponse{
		State: types.ArenaShopState{
			FlashCount:      state.FlashCount,
			NextFlashTime:   state.NextFlashTime,
			LastRefreshTime: state.LastRefreshTime,
		},
		Items: make([]types.ArenaShopItem, 0, len(items)),
	}
	for _, item := range items {
		response.Items = append(response.Items, types.ArenaShopItem{
			ShopID: item.GetShopId(),
			Count:  item.GetCount(),
		})
	}
	return response
}

// PlayerMedalShop godoc
// @Summary     Get player medal shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerMedalShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/medal-shop [get]
func (handler *PlayerHandler) PlayerMedalShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	config, err := medalshop.LoadConfig()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop config", nil))
		return
	}
	state, goods, err := medalshop.RefreshIfNeeded(commander.CommanderID, time.Now(), config)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop state", nil))
		return
	}
	payload := buildMedalShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// RefreshPlayerMedalShop godoc
// @Summary     Refresh player medal shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerMedalShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/medal-shop/refresh [post]
func (handler *PlayerHandler) RefreshPlayerMedalShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	config, err := medalshop.LoadConfig()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop config", nil))
		return
	}
	state, _, err := medalshop.EnsureState(commander.CommanderID, time.Now(), config)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop state", nil))
		return
	}
	goods, err := medalshop.RefreshGoods(commander.CommanderID, config, medalshop.RefreshOptions{
		NextRefreshTime: medalshop.NextDailyReset(time.Now()),
	})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to refresh medal shop", nil))
		return
	}
	if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to reload medal shop state", nil))
		return
	}
	payload := buildMedalShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerMedalShop godoc
// @Summary     Update player medal shop
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id    path  int  true  "Player ID"
// @Param       body  body  types.MedalShopUpdateRequest  true  "State updates"
// @Success     200  {object}  PlayerMedalShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/medal-shop [put]
func (handler *PlayerHandler) UpdatePlayerMedalShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.MedalShopUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	config, err := medalshop.LoadConfig()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop config", nil))
		return
	}
	state, goods, err := medalshop.EnsureState(commander.CommanderID, time.Now(), config)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop state", nil))
		return
	}
	if req.NextRefreshTime != nil {
		if err := orm.GormDB.Model(&orm.MedalShopState{}).
			Where("commander_id = ?", commander.CommanderID).
			Updates(map[string]interface{}{"next_refresh_time": *req.NextRefreshTime}).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update medal shop state", nil))
			return
		}
		if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to reload medal shop state", nil))
			return
		}
	}
	payload := buildMedalShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

func buildMedalShopResponse(state *orm.MedalShopState, goods []orm.MedalShopGood) types.MedalShopResponse {
	response := types.MedalShopResponse{
		State: types.MedalShopState{
			NextRefreshTime: state.NextRefreshTime,
		},
		Items: make([]types.MedalShopItem, 0, len(goods)),
	}
	for _, good := range goods {
		response.Items = append(response.Items, types.MedalShopItem{
			ID:    good.GoodsID,
			Count: good.Count,
			Index: good.Index,
		})
	}
	return response
}

func buildShoppingStreetResponse(state *orm.ShoppingStreetState, goods []orm.ShoppingStreetGood, includeOffers bool) (types.ShoppingStreetResponse, error) {
	lastRefreshedAt := uint32(0)
	if state.NextFlashTime >= shopstreet.DefaultRefreshSeconds {
		lastRefreshedAt = state.NextFlashTime - shopstreet.DefaultRefreshSeconds
	}
	response := types.ShoppingStreetResponse{
		State: types.ShoppingStreetState{
			Level:           state.Level,
			NextFlashTime:   state.NextFlashTime,
			LevelUpTime:     state.LevelUpTime,
			FlashCount:      state.FlashCount,
			LastRefreshedAt: lastRefreshedAt,
		},
		Goods: make([]types.ShoppingStreetGood, 0, len(goods)),
	}
	var offerLookup map[uint32]orm.ShopOffer
	if includeOffers {
		ids := make([]uint32, 0, len(goods))
		for _, good := range goods {
			ids = append(ids, good.GoodsID)
		}
		lookup, err := loadShoppingStreetOffers(ids)
		if err != nil {
			return types.ShoppingStreetResponse{}, err
		}
		offerLookup = lookup
	}
	for _, good := range goods {
		entry := types.ShoppingStreetGood{
			GoodsID:  good.GoodsID,
			Discount: good.Discount,
			BuyCount: good.BuyCount,
		}
		if includeOffers {
			if offer, ok := offerLookup[good.GoodsID]; ok {
				entry.Offer = &types.ShoppingStreetOfferSummary{
					ID:             offer.ID,
					ResourceNumber: offer.ResourceNumber,
					ResourceID:     offer.ResourceID,
					Type:           offer.Type,
					Number:         offer.Number,
					Genre:          offer.Genre,
					Discount:       offer.Discount,
					EffectArgs:     types.RawJSON{Value: offer.EffectArgs},
				}
			}
		}
		response.Goods = append(response.Goods, entry)
	}
	return response, nil
}

func loadShoppingStreetOffers(ids []uint32) (map[uint32]orm.ShopOffer, error) {
	lookup := make(map[uint32]orm.ShopOffer)
	if len(ids) == 0 {
		return lookup, nil
	}
	var offers []orm.ShopOffer
	if err := orm.GormDB.Where("id IN ?", ids).Find(&offers).Error; err != nil {
		return nil, err
	}
	for _, offer := range offers {
		lookup[offer.ID] = offer
	}
	return lookup, nil
}

func (handler *PlayerHandler) playerListResponse(result orm.PlayerListResult, params orm.PlayerQueryParams) (types.PlayerListResponse, error) {
	banLookup, err := buildBanLookup(result.Commanders)
	if err != nil {
		return types.PlayerListResponse{}, err
	}
	onlineIDs := onlineCommanderIDs()
	players := make([]types.PlayerSummary, 0, len(result.Commanders))
	for _, commander := range result.Commanders {
		players = append(players, types.PlayerSummary{
			CommanderID: commander.CommanderID,
			AccountID:   commander.AccountID,
			Name:        commander.Name,
			Level:       commander.Level,
			LastLogin:   commander.LastLogin.UTC().Format(time.RFC3339),
			Banned:      banLookup[commander.CommanderID],
			Online:      onlineIDs[commander.CommanderID],
		})
	}

	return types.PlayerListResponse{
		Players: players,
		Meta: types.PaginationMeta{
			Offset: params.Offset,
			Limit:  params.Limit,
			Total:  result.Total,
		},
	}, nil
}

func buildBanLookup(commanders []orm.Commander) (map[uint32]bool, error) {
	ids := make([]uint32, 0, len(commanders))
	for _, commander := range commanders {
		ids = append(ids, commander.CommanderID)
	}
	if len(ids) == 0 {
		return map[uint32]bool{}, nil
	}

	var punishments []orm.Punishment
	if err := orm.GormDB.
		Where("punished_id IN ? AND lift_timestamp IS NULL", ids).
		Find(&punishments).Error; err != nil {
		return nil, err
	}

	lookup := make(map[uint32]bool, len(punishments))
	for _, punishment := range punishments {
		lookup[punishment.PunishedID] = true
	}
	return lookup, nil
}

func onlineCommanderIDs() map[uint32]bool {
	lookup := make(map[uint32]bool)
	for _, client := range connection.BelfastInstance.ListClients() {
		if client.Commander != nil {
			lookup[client.Commander.CommanderID] = true
		}
	}
	return lookup
}

func onlineIDSlice() []uint32 {
	lookup := onlineCommanderIDs()
	ids := make([]uint32, 0, len(lookup))
	for id := range lookup {
		ids = append(ids, id)
	}
	return ids
}

func dedupeResourceUpdates(entries []types.ResourceUpdateEntry) bool {
	seen := make(map[uint32]struct{}, len(entries))
	for _, entry := range entries {
		if _, exists := seen[entry.ResourceID]; exists {
			return false
		}
		seen[entry.ResourceID] = struct{}{}
	}
	return true
}

func findCommanderClient(commanderID uint32) *connection.Client {
	for _, client := range connection.BelfastInstance.ListClients() {
		if client.Commander != nil && client.Commander.CommanderID == commanderID {
			return client
		}
	}
	return nil
}
