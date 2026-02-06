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
	"gorm.io/gorm/clause"

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
	maxLimit = 200
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
	party.Post("", handler.CreatePlayer)
	party.Patch("/{id:uint}", handler.UpdatePlayer)
	party.Post("/compensations/push-online", handler.PushOnlineCompensationNotifications)
	party.Get("/{id:uint}/resources", handler.PlayerResources)
	party.Get("/{id:uint}/resources/{resource_id:uint}", handler.PlayerResource)
	party.Delete("/{id:uint}/resources/{resource_id:uint}", handler.DeletePlayerResource)
	party.Get("/{id:uint}/ships", handler.PlayerShips)
	party.Get("/{id:uint}/ships/{owned_id:uint}", handler.PlayerShip)
	party.Post("/{id:uint}/ships", handler.CreatePlayerShip)
	party.Patch("/{id:uint}/ships/{owned_id:uint}", handler.UpdatePlayerShip)
	party.Delete("/{id:uint}/ships/{owned_id:uint}", handler.DeletePlayerShip)
	party.Get("/{id:uint}/secretaries", handler.PlayerSecretaries)
	party.Put("/{id:uint}/secretaries", handler.ReplacePlayerSecretaries)
	party.Delete("/{id:uint}/secretaries", handler.DeletePlayerSecretaries)
	party.Get("/{id:uint}/items", handler.PlayerItems)
	party.Get("/{id:uint}/items/{item_id:uint}", handler.PlayerItem)
	party.Patch("/{id:uint}/items/{item_id:uint}", handler.UpdatePlayerItemQuantity)
	party.Delete("/{id:uint}/items/{item_id:uint}", handler.DeletePlayerItem)
	party.Get("/{id:uint}/equipment", handler.PlayerEquipment)
	party.Get("/{id:uint}/equipment/{equipment_id:uint}", handler.PlayerEquipmentEntry)
	party.Post("/{id:uint}/equipment", handler.UpsertPlayerEquipment)
	party.Delete("/{id:uint}/equipment/{equipment_id:uint}", handler.DeletePlayerEquipment)
	party.Get("/{id:uint}/ships/{owned_id:uint}/equipment", handler.PlayerShipEquipment)
	party.Patch("/{id:uint}/ships/{owned_id:uint}/equipment", handler.UpdatePlayerShipEquipment)
	party.Get("/{id:uint}/misc-items", handler.PlayerMiscItems)
	party.Get("/{id:uint}/misc-items/{item_id:uint}", handler.PlayerMiscItem)
	party.Put("/{id:uint}/misc-items/{item_id:uint}", handler.UpdatePlayerMiscItem)
	party.Delete("/{id:uint}/misc-items/{item_id:uint}", handler.DeletePlayerMiscItem)
	party.Get("/{id:uint}/remaster", handler.PlayerRemasterState)
	party.Patch("/{id:uint}/remaster", handler.UpdatePlayerRemasterState)
	party.Get("/{id:uint}/remaster/progress", handler.PlayerRemasterProgress)
	party.Post("/{id:uint}/remaster/progress", handler.UpsertPlayerRemasterProgress)
	party.Patch("/{id:uint}/remaster/progress/{chapter_id:uint}/{pos:uint}", handler.UpdatePlayerRemasterProgress)
	party.Delete("/{id:uint}/remaster/progress/{chapter_id:uint}/{pos:uint}", handler.DeletePlayerRemasterProgress)
	party.Get("/{id:uint}/chapter-state", handler.PlayerChapterState)
	party.Get("/{id:uint}/chapter-state/search", handler.SearchPlayerChapterStates)
	party.Post("/{id:uint}/chapter-state", handler.CreatePlayerChapterState)
	party.Patch("/{id:uint}/chapter-state", handler.UpdatePlayerChapterState)
	party.Delete("/{id:uint}/chapter-state", handler.DeletePlayerChapterState)
	party.Get("/{id:uint}/chapter-progress", handler.ListPlayerChapterProgress)
	party.Get("/{id:uint}/chapter-progress/search", handler.SearchPlayerChapterProgress)
	party.Get("/{id:uint}/chapter-progress/{chapter_id:uint}", handler.PlayerChapterProgress)
	party.Post("/{id:uint}/chapter-progress", handler.CreatePlayerChapterProgress)
	party.Patch("/{id:uint}/chapter-progress/{chapter_id:uint}", handler.UpdatePlayerChapterProgress)
	party.Delete("/{id:uint}/chapter-progress/{chapter_id:uint}", handler.DeletePlayerChapterProgress)
	party.Get("/{id:uint}/builds", handler.PlayerBuilds)
	party.Get("/{id:uint}/builds/{build_id:uint}", handler.PlayerBuild)
	party.Post("/{id:uint}/builds", handler.CreatePlayerBuild)
	party.Get("/{id:uint}/builds/queue", handler.PlayerBuildQueue)
	party.Patch("/{id:uint}/builds/counters", handler.UpdatePlayerBuildCounters)
	party.Get("/{id:uint}/support-requisition", handler.PlayerSupportRequisition)
	party.Post("/{id:uint}/support-requisition/reset", handler.ResetPlayerSupportRequisition)
	party.Patch("/{id:uint}/builds/{build_id:uint}", handler.UpdatePlayerBuild)
	party.Patch("/{id:uint}/builds/{build_id:uint}/quick-finish", handler.QuickFinishBuild)
	party.Delete("/{id:uint}/builds/{build_id:uint}", handler.DeletePlayerBuild)
	party.Get("/{id:uint}/mails", handler.PlayerMails)
	party.Get("/{id:uint}/mails/{mail_id:uint}", handler.PlayerMail)
	party.Patch("/{id:uint}/mails/{mail_id:uint}", handler.UpdatePlayerMail)
	party.Delete("/{id:uint}/mails/{mail_id:uint}", handler.DeletePlayerMail)
	party.Get("/{id:uint}/compensations", handler.PlayerCompensations)
	party.Get("/{id:uint}/compensations/{compensation_id:uint}", handler.PlayerCompensation)
	party.Post("/{id:uint}/compensations/push", handler.PushCompensationNotification)
	party.Get("/{id:uint}/tb", handler.PlayerTB)
	party.Post("/{id:uint}/tb", handler.CreatePlayerTB)
	party.Put("/{id:uint}/tb", handler.UpdatePlayerTB)
	party.Delete("/{id:uint}/tb", handler.DeletePlayerTB)
	party.Get("/{id:uint}/fleets", handler.PlayerFleets)
	party.Get("/{id:uint}/fleets/{fleet_id:uint}", handler.PlayerFleet)
	party.Post("/{id:uint}/fleets", handler.CreatePlayerFleet)
	party.Patch("/{id:uint}/fleets/{fleet_id:uint}", handler.UpdatePlayerFleet)
	party.Delete("/{id:uint}/fleets/{fleet_id:uint}", handler.DeletePlayerFleet)
	party.Get("/{id:uint}/skins", handler.PlayerSkins)
	party.Get("/{id:uint}/skins/{skin_id:uint}", handler.PlayerSkin)
	party.Patch("/{id:uint}/skins/{skin_id:uint}", handler.UpdatePlayerSkin)
	party.Delete("/{id:uint}/skins/{skin_id:uint}", handler.DeletePlayerSkin)
	party.Get("/{id:uint}/buffs", handler.PlayerBuffs)
	party.Get("/{id:uint}/buffs/{buff_id:uint}", handler.PlayerBuff)
	party.Get("/{id:uint}/flags", handler.PlayerFlags)
	party.Post("/{id:uint}/flags", handler.AddPlayerFlag)
	party.Delete("/{id:uint}/flags/{flag_id:uint}", handler.DeletePlayerFlag)
	party.Get("/{id:uint}/likes", handler.PlayerLikes)
	party.Post("/{id:uint}/likes", handler.AddPlayerLike)
	party.Delete("/{id:uint}/likes/{group_id:uint}", handler.DeletePlayerLike)
	party.Get("/{id:uint}/random-flagships", handler.PlayerRandomFlagShips)
	party.Post("/{id:uint}/random-flagships", handler.UpsertPlayerRandomFlagShip)
	party.Delete("/{id:uint}/random-flagships/{ship_id:uint}/{phantom_id:uint}", handler.DeletePlayerRandomFlagShip)
	party.Get("/{id:uint}/random-flagship", handler.PlayerRandomFlagShip)
	party.Patch("/{id:uint}/random-flagship", handler.UpdatePlayerRandomFlagShip)
	party.Get("/{id:uint}/random-flagship-mode", handler.PlayerRandomFlagShipMode)
	party.Patch("/{id:uint}/random-flagship-mode", handler.UpdatePlayerRandomFlagShipMode)
	party.Get("/{id:uint}/guide", handler.PlayerGuide)
	party.Patch("/{id:uint}/guide", handler.UpdatePlayerGuide)
	party.Get("/{id:uint}/stories", handler.PlayerStories)
	party.Post("/{id:uint}/stories", handler.AddPlayerStory)
	party.Put("/{id:uint}/stories/{story_id:uint}", handler.UpsertPlayerStory)
	party.Delete("/{id:uint}/stories/{story_id:uint}", handler.DeletePlayerStory)
	party.Get("/{id:uint}/attires", handler.PlayerAttires)
	party.Post("/{id:uint}/attires", handler.AddPlayerAttire)
	party.Patch("/{id:uint}/attires/{type:uint}/{attire_id:uint}", handler.UpdatePlayerAttire)
	party.Delete("/{id:uint}/attires/{type:uint}/{attire_id:uint}", handler.DeletePlayerAttire)
	party.Patch("/{id:uint}/attires/selected", handler.UpdatePlayerAttireSelection)
	party.Get("/{id:uint}/livingarea-covers", handler.PlayerLivingAreaCovers)
	party.Post("/{id:uint}/livingarea-covers", handler.AddPlayerLivingAreaCover)
	party.Patch("/{id:uint}/livingarea-covers/{cover_id:uint}", handler.UpdatePlayerLivingAreaCoverState)
	party.Delete("/{id:uint}/livingarea-covers/{cover_id:uint}", handler.DeletePlayerLivingAreaCover)
	party.Patch("/{id:uint}/livingarea-covers/selected", handler.UpdatePlayerLivingAreaCover)
	party.Get("/{id:uint}/shopping-street", handler.PlayerShoppingStreet)
	party.Post("/{id:uint}/shopping-street/refresh", handler.RefreshPlayerShoppingStreet)
	party.Put("/{id:uint}/shopping-street", handler.UpdatePlayerShoppingStreet)
	party.Delete("/{id:uint}/shopping-street", handler.DeletePlayerShoppingStreet)
	party.Get("/{id:uint}/shopping-street/goods", handler.PlayerShoppingStreetGoods)
	party.Get("/{id:uint}/shopping-street/goods/{goods_id:uint}", handler.PlayerShoppingStreetGood)
	party.Post("/{id:uint}/shopping-street/goods", handler.AddPlayerShoppingStreetGood)
	party.Put("/{id:uint}/shopping-street/goods", handler.ReplacePlayerShoppingStreetGoods)
	party.Patch("/{id:uint}/shopping-street/goods/{goods_id:uint}", handler.UpdatePlayerShoppingStreetGood)
	party.Delete("/{id:uint}/shopping-street/goods/{goods_id:uint}", handler.DeletePlayerShoppingStreetGood)
	party.Get("/{id:uint}/arena-shop", handler.PlayerArenaShop)
	party.Post("/{id:uint}/arena-shop/refresh", handler.RefreshPlayerArenaShop)
	party.Put("/{id:uint}/arena-shop", handler.UpdatePlayerArenaShop)
	party.Delete("/{id:uint}/arena-shop", handler.DeletePlayerArenaShop)
	party.Get("/{id:uint}/medal-shop", handler.PlayerMedalShop)
	party.Post("/{id:uint}/medal-shop/refresh", handler.RefreshPlayerMedalShop)
	party.Put("/{id:uint}/medal-shop", handler.UpdatePlayerMedalShop)
	party.Get("/{id:uint}/medal-shop/goods", handler.PlayerMedalShopGoods)
	party.Post("/{id:uint}/medal-shop/goods", handler.AddPlayerMedalShopGood)
	party.Patch("/{id:uint}/medal-shop/goods/{index:uint}", handler.UpdatePlayerMedalShopGood)
	party.Delete("/{id:uint}/medal-shop/goods/{index:uint}", handler.DeletePlayerMedalShopGood)
	party.Get("/{id:uint}/guild-shop", handler.PlayerGuildShop)
	party.Post("/{id:uint}/guild-shop/refresh", handler.RefreshPlayerGuildShop)
	party.Put("/{id:uint}/guild-shop", handler.UpdatePlayerGuildShop)
	party.Get("/{id:uint}/guild-shop/goods", handler.PlayerGuildShopGoods)
	party.Post("/{id:uint}/guild-shop/goods", handler.AddPlayerGuildShopGood)
	party.Patch("/{id:uint}/guild-shop/goods/{index:uint}", handler.UpdatePlayerGuildShopGood)
	party.Delete("/{id:uint}/guild-shop/goods/{index:uint}", handler.DeletePlayerGuildShopGood)
	party.Get("/{id:uint}/minigame-shop", handler.PlayerMiniGameShop)
	party.Post("/{id:uint}/minigame-shop/refresh", handler.RefreshPlayerMiniGameShop)
	party.Put("/{id:uint}/minigame-shop", handler.UpdatePlayerMiniGameShop)
	party.Get("/{id:uint}/minigame-shop/goods", handler.PlayerMiniGameShopGoods)
	party.Post("/{id:uint}/minigame-shop/goods", handler.AddPlayerMiniGameShopGood)
	party.Patch("/{id:uint}/minigame-shop/goods/{goods_id:uint}", handler.UpdatePlayerMiniGameShopGood)
	party.Delete("/{id:uint}/minigame-shop/goods/{goods_id:uint}", handler.DeletePlayerMiniGameShopGood)
	party.Get("/{id:uint}/punishments", handler.PlayerPunishments)
	party.Get("/{id:uint}/punishments/{punishment_id:uint}", handler.PlayerPunishment)
	party.Post("/{id:uint}/punishments", handler.CreatePlayerPunishment)
	party.Patch("/{id:uint}/punishments/{punishment_id:uint}", handler.UpdatePlayerPunishment)
	party.Delete("/{id:uint}/punishments/{punishment_id:uint}", handler.DeletePlayerPunishment)
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
	party.Patch("/{id:uint}/buffs/{buff_id:uint}", handler.UpdatePlayerBuff)
	party.Delete("/{id:uint}/buffs/{buff_id:uint}", handler.DeletePlayerBuff)
	party.Delete("/{id:uint}", handler.DeletePlayer)
}

func parsePagination(ctx iris.Context) (types.PaginationMeta, error) {
	offset, err := parseQueryInt(ctx.URLParamDefault("offset", "0"))
	if err != nil || offset < 0 {
		return types.PaginationMeta{}, fmt.Errorf("offset must be >= 0")
	}
	limitValue := strings.TrimSpace(ctx.URLParam("limit"))
	if limitValue == "" {
		return types.PaginationMeta{Offset: offset, Limit: 0}, nil
	}
	limit, err := parseQueryInt(limitValue)
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

// SearchPlayers godoc
// @Summary     Search players
// @Tags        Players
// @Produce     json
// @Param       q         query  string  true   "Search query"
// @Param       offset    query  int     false  "Pagination offset"
// @Param       limit     query  int     false  "Pagination limit"
// @Param       sort      query  string  false  "Sort by last_login"
// @Param       filter    query  string  false  "Filters: online, banned"
// @Param       min_level query  int     false  "Minimum level"
// @Success     200  {object}  ListPlayersResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/search [get]
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

// CreatePlayer godoc
// @Summary     Create player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       payload  body  types.PlayerCreateRequest  true  "Player create"
// @Success     200  {object}  PlayerMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players [post]
func (handler *PlayerHandler) CreatePlayer(ctx iris.Context) {
	var req types.PlayerCreateRequest
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

	name := strings.TrimSpace(req.Name)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	var existing orm.Commander
	if err := orm.GormDB.Select("commander_id").Where("commander_id = ?", req.CommanderID).First(&existing).Error; err == nil {
		ctx.StatusCode(iris.StatusConflict)
		_ = ctx.JSON(response.Error("conflict", "commander already exists", nil))
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to check commander", nil))
		return
	}
	if err := orm.GormDB.Select("commander_id").Where("name = ?", name).First(&existing).Error; err == nil {
		ctx.StatusCode(iris.StatusConflict)
		_ = ctx.JSON(response.Error("conflict", "name already exists", nil))
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to check name", nil))
		return
	}

	commander := orm.Commander{
		CommanderID: req.CommanderID,
		AccountID:   req.AccountID,
		Name:        name,
	}
	if req.Level != nil {
		commander.Level = *req.Level
	}
	if req.Exp != nil {
		commander.Exp = *req.Exp
	}
	if req.LastLogin != nil {
		parsed, err := time.Parse(time.RFC3339, *req.LastLogin)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "last_login must be RFC3339", nil))
			return
		}
		commander.LastLogin = parsed
	}
	if req.GuideIndex != nil {
		commander.GuideIndex = *req.GuideIndex
	}
	if req.NewGuideIndex != nil {
		commander.NewGuideIndex = *req.NewGuideIndex
	}
	if req.NameChangeCooldown != nil {
		parsed, err := time.Parse(time.RFC3339, *req.NameChangeCooldown)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "name_change_cooldown must be RFC3339", nil))
			return
		}
		commander.NameChangeCooldown = parsed
	}
	if req.RoomID != nil {
		commander.RoomID = *req.RoomID
	}
	if req.ExchangeCount != nil {
		commander.ExchangeCount = *req.ExchangeCount
	}
	if req.DrawCount1 != nil {
		commander.DrawCount1 = *req.DrawCount1
	}
	if req.DrawCount10 != nil {
		commander.DrawCount10 = *req.DrawCount10
	}
	if req.SupportRequisitionCount != nil {
		commander.SupportRequisitionCount = *req.SupportRequisitionCount
	}
	if req.SupportRequisitionMonth != nil {
		commander.SupportRequisitionMonth = *req.SupportRequisitionMonth
	}
	if req.AccPayLv != nil {
		commander.AccPayLv = *req.AccPayLv
	}
	if req.LivingAreaCoverID != nil {
		commander.LivingAreaCoverID = *req.LivingAreaCoverID
	}
	if req.SelectedIconFrameID != nil {
		commander.SelectedIconFrameID = *req.SelectedIconFrameID
	}
	if req.SelectedChatFrameID != nil {
		commander.SelectedChatFrameID = *req.SelectedChatFrameID
	}
	if req.SelectedBattleUIID != nil {
		commander.SelectedBattleUIID = *req.SelectedBattleUIID
	}
	if req.DisplayIconID != nil {
		commander.DisplayIconID = *req.DisplayIconID
	}
	if req.DisplaySkinID != nil {
		commander.DisplaySkinID = *req.DisplaySkinID
	}
	if req.DisplayIconThemeID != nil {
		commander.DisplayIconThemeID = *req.DisplayIconThemeID
	}
	if req.RandomShipMode != nil {
		commander.RandomShipMode = *req.RandomShipMode
	}
	if req.RandomFlagShipEnabled != nil {
		commander.RandomFlagShipEnabled = *req.RandomFlagShipEnabled
	}

	if err := orm.GormDB.Create(&commander).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "commander already exists", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create player", nil))
		return
	}

	if err := orm.GormDB.First(&commander, req.CommanderID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load player", nil))
		return
	}
	banStatus, err := orm.GetBanStatus(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load ban status", nil))
		return
	}
	onlineIDs := onlineCommanderIDs()
	payload := types.PlayerMutationResponse{
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

// UpdatePlayer godoc
// @Summary     Update player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int                     true  "Player ID"
// @Param       payload  body  types.PlayerUpdateRequest  true  "Player updates"
// @Success     200  {object}  PlayerMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id} [patch]
func (handler *PlayerHandler) UpdatePlayer(ctx iris.Context) {
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

	var req types.PlayerUpdateRequest
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
	if req.AccountID != nil {
		updates["account_id"] = *req.AccountID
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
			return
		}
		var existing orm.Commander
		if err := orm.GormDB.Select("commander_id").Where("name = ? AND commander_id <> ?", name, commanderID).First(&existing).Error; err == nil {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "name already exists", nil))
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to check name", nil))
			return
		}
		updates["name"] = name
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if req.Exp != nil {
		updates["exp"] = *req.Exp
	}
	if req.LastLogin != nil {
		parsed, err := time.Parse(time.RFC3339, *req.LastLogin)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "last_login must be RFC3339", nil))
			return
		}
		updates["last_login"] = parsed
	}
	if req.GuideIndex != nil {
		updates["guide_index"] = *req.GuideIndex
	}
	if req.NewGuideIndex != nil {
		updates["new_guide_index"] = *req.NewGuideIndex
	}
	if req.NameChangeCooldown != nil {
		parsed, err := time.Parse(time.RFC3339, *req.NameChangeCooldown)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "name_change_cooldown must be RFC3339", nil))
			return
		}
		updates["name_change_cooldown"] = parsed
	}
	if req.RoomID != nil {
		updates["room_id"] = *req.RoomID
	}
	if req.ExchangeCount != nil {
		updates["exchange_count"] = *req.ExchangeCount
	}
	if req.DrawCount1 != nil {
		updates["draw_count1"] = *req.DrawCount1
	}
	if req.DrawCount10 != nil {
		updates["draw_count10"] = *req.DrawCount10
	}
	if req.SupportRequisitionCount != nil {
		updates["support_requisition_count"] = *req.SupportRequisitionCount
	}
	if req.SupportRequisitionMonth != nil {
		updates["support_requisition_month"] = *req.SupportRequisitionMonth
	}
	if req.AccPayLv != nil {
		updates["acc_pay_lv"] = *req.AccPayLv
	}
	if req.LivingAreaCoverID != nil {
		updates["living_area_cover_id"] = *req.LivingAreaCoverID
	}
	if req.SelectedIconFrameID != nil {
		updates["selected_icon_frame_id"] = *req.SelectedIconFrameID
	}
	if req.SelectedChatFrameID != nil {
		updates["selected_chat_frame_id"] = *req.SelectedChatFrameID
	}
	if req.SelectedBattleUIID != nil {
		updates["selected_battle_ui_id"] = *req.SelectedBattleUIID
	}
	if req.DisplayIconID != nil {
		updates["display_icon_id"] = *req.DisplayIconID
	}
	if req.DisplaySkinID != nil {
		updates["display_skin_id"] = *req.DisplaySkinID
	}
	if req.DisplayIconThemeID != nil {
		updates["display_icon_theme_id"] = *req.DisplayIconThemeID
	}
	if req.RandomShipMode != nil {
		updates["random_ship_mode"] = *req.RandomShipMode
	}
	if req.RandomFlagShipEnabled != nil {
		updates["random_flag_ship_enabled"] = *req.RandomFlagShipEnabled
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}

	if err := orm.GormDB.Model(&commander).Updates(updates).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "name already exists", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update player", nil))
		return
	}

	if err := orm.GormDB.First(&commander, commanderID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load player", nil))
		return
	}
	banStatus, err := orm.GetBanStatus(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load ban status", nil))
		return
	}
	onlineIDs := onlineCommanderIDs()
	payload := types.PlayerMutationResponse{
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

// PlayerResource godoc
// @Summary     Get player resource
// @Tags        Players
// @Produce     json
// @Param       id           path  int  true  "Player ID"
// @Param       resource_id  path  int  true  "Resource ID"
// @Success     200  {object}  PlayerResourceEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/resources/{resource_id} [get]
func (handler *PlayerHandler) PlayerResource(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	resourceID, err := parsePathUint32(ctx.Params().Get("resource_id"), "resource id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	owned, ok := commander.OwnedResourcesMap[resourceID]
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "resource not owned", nil))
		return
	}

	payload := types.PlayerResourceEntry{
		ResourceID: owned.ResourceID,
		Amount:     owned.Amount,
		Name:       owned.Resource.Name,
	}

	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerResource godoc
// @Summary     Remove player resource
// @Tags        Players
// @Produce     json
// @Param       id           path  int  true  "Player ID"
// @Param       resource_id  path  int  true  "Resource ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/resources/{resource_id} [delete]
func (handler *PlayerHandler) DeletePlayerResource(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	resourceID, err := parsePathUint32(ctx.Params().Get("resource_id"), "resource id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if commander.OwnedResourcesMap[resourceID] == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "resource not owned", nil))
		return
	}

	if err := orm.GormDB.Where("commander_id = ? AND resource_id = ?", commander.CommanderID, resourceID).Delete(&orm.OwnedResource{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete resource", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
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

// PlayerItem godoc
// @Summary     Get player item
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       item_id  path  int  true  "Item ID"
// @Success     200  {object}  PlayerItemEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/items/{item_id} [get]
func (handler *PlayerHandler) PlayerItem(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	itemID, err := parsePathUint32(ctx.Params().Get("item_id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	commanderItem, hasCommanderItem := commander.CommanderItemsMap[itemID]
	miscItem, hasMiscItem := commander.MiscItemsMap[itemID]
	if !hasCommanderItem && !hasMiscItem {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "item not owned", nil))
		return
	}

	count := uint32(0)
	name := ""
	if hasCommanderItem {
		count += commanderItem.Count
		name = commanderItem.Item.Name
	}
	if hasMiscItem {
		count += miscItem.Data
		if name == "" {
			name = miscItem.Item.Name
		}
	}

	payload := types.PlayerItemEntry{
		ItemID: itemID,
		Count:  count,
		Name:   name,
	}

	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerItemQuantity godoc
// @Summary     Update player item quantity
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       item_id   path  int  true  "Item ID"
// @Param       payload  body  types.PlayerItemQuantityUpdateRequest  true  "Item quantity update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/items/{item_id} [patch]
func (handler *PlayerHandler) UpdatePlayerItemQuantity(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	itemID, err := parsePathUint32(ctx.Params().Get("item_id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.PlayerItemQuantityUpdateRequest
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

	if err := commander.SetItem(itemID, req.Quantity); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update item", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerItem godoc
// @Summary     Remove player item
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       item_id  path  int  true  "Item ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/items/{item_id} [delete]
func (handler *PlayerHandler) DeletePlayerItem(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	itemID, err := parsePathUint32(ctx.Params().Get("item_id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	_, hasCommanderItem := commander.CommanderItemsMap[itemID]
	_, hasMiscItem := commander.MiscItemsMap[itemID]
	if !hasCommanderItem && !hasMiscItem {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "item not owned", nil))
		return
	}

	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if hasCommanderItem {
			if err := tx.Where("commander_id = ? AND item_id = ?", commander.CommanderID, itemID).Delete(&orm.CommanderItem{}).Error; err != nil {
				return err
			}
		}
		if hasMiscItem {
			if err := tx.Where("commander_id = ? AND item_id = ?", commander.CommanderID, itemID).Delete(&orm.CommanderMiscItem{}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete item", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// PlayerEquipment godoc
// @Summary     Get player equipment bag
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerEquipmentResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/equipment [get]
func (handler *PlayerHandler) PlayerEquipment(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	payload := types.PlayerEquipmentResponse{Equipment: make([]types.PlayerEquipmentEntry, 0, len(commander.OwnedEquipments))}
	for _, entry := range commander.OwnedEquipments {
		payload.Equipment = append(payload.Equipment, types.PlayerEquipmentEntry{
			EquipmentID: entry.EquipmentID,
			Count:       entry.Count,
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

// PlayerEquipmentEntry godoc
// @Summary     Get player equipment entry
// @Tags        Players
// @Produce     json
// @Param       id            path  int  true  "Player ID"
// @Param       equipment_id  path  int  true  "Equipment ID"
// @Success     200  {object}  PlayerEquipmentEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/equipment/{equipment_id} [get]
func (handler *PlayerHandler) PlayerEquipmentEntry(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	equipmentID, err := parsePathUint32(ctx.Params().Get("equipment_id"), "equipment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	entry := commander.GetOwnedEquipment(equipmentID)
	if entry == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "equipment not owned", nil))
		return
	}
	payload := types.PlayerEquipmentEntry{EquipmentID: entry.EquipmentID, Count: entry.Count}
	_ = ctx.JSON(response.Success(payload))
}

// UpsertPlayerEquipment godoc
// @Summary     Upsert player equipment entry
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       payload  body  types.PlayerEquipmentUpsertRequest  true  "Equipment upsert"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/equipment [post]
func (handler *PlayerHandler) UpsertPlayerEquipment(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerEquipmentUpsertRequest
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
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		return commander.SetOwnedEquipmentTx(tx, req.EquipmentID, req.Count)
	}); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update equipment", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerEquipment godoc
// @Summary     Remove player equipment entry
// @Tags        Players
// @Produce     json
// @Param       id            path  int  true  "Player ID"
// @Param       equipment_id  path  int  true  "Equipment ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/equipment/{equipment_id} [delete]
func (handler *PlayerHandler) DeletePlayerEquipment(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	equipmentID, err := parsePathUint32(ctx.Params().Get("equipment_id"), "equipment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if commander.GetOwnedEquipment(equipmentID) == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "equipment not owned", nil))
		return
	}
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		return commander.SetOwnedEquipmentTx(tx, equipmentID, 0)
	}); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete equipment", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerShipEquipment godoc
// @Summary     Get player ship equipment slots
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       owned_id  path  int  true  "Owned ship ID"
// @Success     200  {object}  PlayerShipEquipmentResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ships/{owned_id}/equipment [get]
func (handler *PlayerHandler) PlayerShipEquipment(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	ownedID, err := parsePathUint32(ctx.Params().Get("owned_id"), "owned id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	ship, ok := commander.OwnedShipsMap[ownedID]
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "ship not owned", nil))
		return
	}
	entries, err := orm.ListOwnedShipEquipment(orm.GormDB, commander.CommanderID, ship.ID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load ship equipment", nil))
		return
	}
	payload := types.PlayerShipEquipmentResponse{Equipment: make([]types.PlayerShipEquipmentEntry, 0, len(entries))}
	for _, entry := range entries {
		payload.Equipment = append(payload.Equipment, types.PlayerShipEquipmentEntry{
			Pos:     entry.Pos,
			EquipID: entry.EquipID,
			SkinID:  entry.SkinID,
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerShipEquipment godoc
// @Summary     Update player ship equipment slots
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       owned_id  path  int  true  "Owned ship ID"
// @Param       payload   body  types.PlayerShipEquipmentUpdateRequest  true  "Ship equipment update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ships/{owned_id}/equipment [patch]
func (handler *PlayerHandler) UpdatePlayerShipEquipment(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	ownedID, err := parsePathUint32(ctx.Params().Get("owned_id"), "owned id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	ship, ok := commander.OwnedShipsMap[ownedID]
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "ship not owned", nil))
		return
	}
	var req types.PlayerShipEquipmentUpdateRequest
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
	config, err := orm.GetShipEquipConfig(ship.ShipID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load ship config", nil))
		return
	}
	maxPos := config.SlotCount()
	for _, entry := range req.Equipment {
		if entry.Pos == 0 || entry.Pos > maxPos {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid equipment position", nil))
			return
		}
	}
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		for _, entry := range req.Equipment {
			update := orm.OwnedShipEquipment{
				OwnerID: commander.CommanderID,
				ShipID:  ship.ID,
				Pos:     entry.Pos,
				EquipID: entry.EquipID,
				SkinID:  entry.SkinID,
			}
			if err := orm.UpsertOwnedShipEquipmentTx(tx, &update); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update ship equipment", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerMiscItems godoc
// @Summary     Get player misc items
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerMiscItemsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/misc-items [get]
func (handler *PlayerHandler) PlayerMiscItems(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerMiscItemResponse{Items: make([]types.PlayerMiscItemEntry, 0, len(commander.MiscItems))}
	for _, miscItem := range commander.MiscItems {
		payload.Items = append(payload.Items, types.PlayerMiscItemEntry{
			ItemID: miscItem.ItemID,
			Data:   miscItem.Data,
			Name:   miscItem.Item.Name,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerMiscItem godoc
// @Summary     Get player misc item
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       item_id  path  int  true  "Item ID"
// @Success     200  {object}  PlayerMiscItemEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/misc-items/{item_id} [get]
func (handler *PlayerHandler) PlayerMiscItem(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	itemID, err := parsePathUint32(ctx.Params().Get("item_id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	item, ok := commander.MiscItemsMap[itemID]
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "misc item not owned", nil))
		return
	}

	payload := types.PlayerMiscItemEntry{
		ItemID: item.ItemID,
		Data:   item.Data,
		Name:   item.Item.Name,
	}

	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerMiscItem godoc
// @Summary     Update player misc item
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       item_id  path  int  true  "Item ID"
// @Param       payload  body  types.PlayerMiscItemUpdateRequest  true  "Misc item update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/misc-items/{item_id} [put]
func (handler *PlayerHandler) UpdatePlayerMiscItem(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	itemID, err := parsePathUint32(ctx.Params().Get("item_id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.PlayerMiscItemUpdateRequest
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

	data := *req.Data
	if _, ok := commander.MiscItemsMap[itemID]; ok {
		if err := orm.GormDB.Model(&orm.CommanderMiscItem{}).
			Where("commander_id = ? AND item_id = ?", commander.CommanderID, itemID).
			Update("data", data).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update misc item", nil))
			return
		}
		_ = ctx.JSON(response.Success(nil))
		return
	}

	newItem := orm.CommanderMiscItem{
		CommanderID: commander.CommanderID,
		ItemID:      itemID,
		Data:        data,
	}
	if err := orm.GormDB.Create(&newItem).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update misc item", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerMiscItem godoc
// @Summary     Remove player misc item
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       item_id  path  int  true  "Item ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/misc-items/{item_id} [delete]
func (handler *PlayerHandler) DeletePlayerMiscItem(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	itemID, err := parsePathUint32(ctx.Params().Get("item_id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if commander.MiscItemsMap[itemID] == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "misc item not owned", nil))
		return
	}

	if err := orm.GormDB.Where("commander_id = ? AND item_id = ?", commander.CommanderID, itemID).Delete(&orm.CommanderMiscItem{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete misc item", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
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

// PlayerSecretaries godoc
// @Summary     Get player secretaries
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerSecretariesResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/secretaries [get]
func (handler *PlayerHandler) PlayerSecretaries(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerSecretariesResponse{Ships: make([]types.PlayerSecretaryEntry, 0, len(commander.Ships))}
	for _, ship := range commander.Ships {
		position := ship.SecretaryPosition
		if !ship.IsSecretary {
			position = nil
		}
		payload.Ships = append(payload.Ships, types.PlayerSecretaryEntry{
			ShipID:      ship.ID,
			PhantomID:   ship.SecretaryPhantomID,
			IsSecretary: ship.IsSecretary,
			Position:    position,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// ReplacePlayerSecretaries godoc
// @Summary     Replace player secretaries
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       payload  body  types.PlayerSecretariesReplaceRequest  true  "Secretary replace"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/secretaries [put]
func (handler *PlayerHandler) ReplacePlayerSecretaries(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.PlayerSecretariesReplaceRequest
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

	updates := make([]orm.SecretaryUpdate, len(req.Secretaries))
	for i, secretary := range req.Secretaries {
		if commander.OwnedShipsMap[secretary.ShipID] == nil {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "ship not owned", nil))
			return
		}
		updates[i] = orm.SecretaryUpdate{
			ShipID:    secretary.ShipID,
			PhantomID: secretary.PhantomID,
		}
	}

	if err := commander.UpdateSecretaries(updates); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update secretaries", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerSecretaries godoc
// @Summary     Remove player secretaries
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/secretaries [delete]
func (handler *PlayerHandler) DeletePlayerSecretaries(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	if err := commander.RemoveSecretaries(); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete secretaries", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// PlayerShip godoc
// @Summary     Get player ship
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       owned_id  path  int  true  "Owned ship ID"
// @Success     200  {object}  PlayerOwnedShipEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ships/{owned_id} [get]
func (handler *PlayerHandler) PlayerShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	ownedID, err := parsePathUint32(ctx.Params().Get("owned_id"), "owned id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	owned, ok := commander.OwnedShipsMap[ownedID]
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "ship not owned", nil))
		return
	}

	payload := buildOwnedShipEntry(*owned)
	_ = ctx.JSON(response.Success(payload))
}

// CreatePlayerShip godoc
// @Summary     Create player ship
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       payload  body  types.PlayerShipCreateRequest  true  "Owned ship create"
// @Success     200  {object}  PlayerOwnedShipEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ships [post]
func (handler *PlayerHandler) CreatePlayerShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.PlayerShipCreateRequest
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
	if err := orm.ValidateShipID(req.ShipID); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid ship_id", nil))
		return
	}

	owned := orm.OwnedShip{
		OwnerID: commander.CommanderID,
		ShipID:  req.ShipID,
	}
	if req.Level != nil {
		owned.Level = *req.Level
	}
	if req.Exp != nil {
		owned.Exp = *req.Exp
	}
	if req.SkinID != nil {
		owned.SkinID = *req.SkinID
	}
	if req.IsLocked != nil {
		owned.IsLocked = *req.IsLocked
	}
	if req.CustomName != nil {
		owned.CustomName = *req.CustomName
	}
	if req.IsSecretary != nil {
		owned.IsSecretary = *req.IsSecretary
	}
	if req.SecretaryPosition != nil {
		owned.SecretaryPosition = req.SecretaryPosition
	}
	if req.SecretaryPhantomID != nil {
		owned.SecretaryPhantomID = *req.SecretaryPhantomID
	}

	if err := orm.GormDB.Create(&owned).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create ship", nil))
		return
	}

	var saved orm.OwnedShip
	if err := orm.GormDB.First(&saved, "id = ? AND owner_id = ?", owned.ID, commander.CommanderID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load ship", nil))
		return
	}

	payload := buildOwnedShipEntry(saved)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerShip godoc
// @Summary     Update player ship
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       owned_id  path  int  true  "Owned ship ID"
// @Param       payload   body  types.PlayerShipUpdateRequest  true  "Owned ship updates"
// @Success     200  {object}  PlayerOwnedShipEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ships/{owned_id} [patch]
func (handler *PlayerHandler) UpdatePlayerShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	ownedID, err := parsePathUint32(ctx.Params().Get("owned_id"), "owned id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if commander.OwnedShipsMap[ownedID] == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "ship not owned", nil))
		return
	}

	var req types.PlayerShipUpdateRequest
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
	if req.ShipID != nil {
		if err := orm.ValidateShipID(*req.ShipID); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid ship_id", nil))
			return
		}
		updates["ship_id"] = *req.ShipID
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if req.Exp != nil {
		updates["exp"] = *req.Exp
	}
	if req.SurplusExp != nil {
		updates["surplus_exp"] = *req.SurplusExp
	}
	if req.MaxLevel != nil {
		updates["max_level"] = *req.MaxLevel
	}
	if req.Intimacy != nil {
		updates["intimacy"] = *req.Intimacy
	}
	if req.IsLocked != nil {
		updates["is_locked"] = *req.IsLocked
	}
	if req.Propose != nil {
		updates["propose"] = *req.Propose
	}
	if req.CommonFlag != nil {
		updates["common_flag"] = *req.CommonFlag
	}
	if req.BlueprintFlag != nil {
		updates["blueprint_flag"] = *req.BlueprintFlag
	}
	if req.Proficiency != nil {
		updates["proficiency"] = *req.Proficiency
	}
	if req.ActivityNPC != nil {
		updates["activity_npc"] = *req.ActivityNPC
	}
	if req.CustomName != nil {
		updates["custom_name"] = *req.CustomName
	}
	if req.ChangeNameTimestamp != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ChangeNameTimestamp)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "change_name_timestamp must be RFC3339", nil))
			return
		}
		updates["change_name_timestamp"] = parsed
	}
	if req.CreateTime != nil {
		parsed, err := time.Parse(time.RFC3339, *req.CreateTime)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "create_time must be RFC3339", nil))
			return
		}
		updates["create_time"] = parsed
	}
	if req.Energy != nil {
		updates["energy"] = *req.Energy
	}
	if req.SkinID != nil {
		updates["skin_id"] = *req.SkinID
	}
	if req.IsSecretary != nil {
		updates["is_secretary"] = *req.IsSecretary
	}
	if req.SecretaryPosition != nil {
		updates["secretary_position"] = *req.SecretaryPosition
	}
	if req.SecretaryPhantomID != nil {
		updates["secretary_phantom_id"] = *req.SecretaryPhantomID
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}

	if err := orm.GormDB.Model(&orm.OwnedShip{}).
		Where("id = ? AND owner_id = ?", ownedID, commander.CommanderID).
		Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update ship", nil))
		return
	}

	var saved orm.OwnedShip
	if err := orm.GormDB.First(&saved, "id = ? AND owner_id = ?", ownedID, commander.CommanderID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load ship", nil))
		return
	}

	payload := buildOwnedShipEntry(saved)
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerShip godoc
// @Summary     Remove player ship
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       owned_id  path  int  true  "Owned ship ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ships/{owned_id} [delete]
func (handler *PlayerHandler) DeletePlayerShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	ownedID, err := parsePathUint32(ctx.Params().Get("owned_id"), "owned id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if commander.OwnedShipsMap[ownedID] == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "ship not owned", nil))
		return
	}

	if err := orm.GormDB.Where("id = ? AND owner_id = ?", ownedID, commander.CommanderID).Delete(&orm.OwnedShip{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete ship", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

func buildOwnedShipEntry(owned orm.OwnedShip) types.PlayerOwnedShipEntry {
	changeNameTimestamp := ""
	if !owned.ChangeNameTimestamp.IsZero() {
		changeNameTimestamp = owned.ChangeNameTimestamp.UTC().Format(time.RFC3339)
	}
	createTime := ""
	if !owned.CreateTime.IsZero() {
		createTime = owned.CreateTime.UTC().Format(time.RFC3339)
	}
	return types.PlayerOwnedShipEntry{
		OwnedID:             owned.ID,
		ShipID:              owned.ShipID,
		Level:               owned.Level,
		Exp:                 owned.Exp,
		SurplusExp:          owned.SurplusExp,
		MaxLevel:            owned.MaxLevel,
		Intimacy:            owned.Intimacy,
		IsLocked:            owned.IsLocked,
		Propose:             owned.Propose,
		CommonFlag:          owned.CommonFlag,
		BlueprintFlag:       owned.BlueprintFlag,
		Proficiency:         owned.Proficiency,
		ActivityNPC:         owned.ActivityNPC,
		CustomName:          owned.CustomName,
		ChangeNameTimestamp: changeNameTimestamp,
		CreateTime:          createTime,
		Energy:              owned.Energy,
		SkinID:              owned.SkinID,
		IsSecretary:         owned.IsSecretary,
		SecretaryPosition:   owned.SecretaryPosition,
		SecretaryPhantomID:  owned.SecretaryPhantomID,
	}
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

// PlayerBuild godoc
// @Summary     Get player build
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       build_id  path  int  true  "Build ID"
// @Success     200  {object}  PlayerBuildEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds/{build_id} [get]
func (handler *PlayerHandler) PlayerBuild(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	buildID, err := parsePathUint32(ctx.Params().Get("build_id"), "build id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	build := orm.Build{ID: buildID}
	if err := build.Retrieve(true); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "build not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load build", nil))
		return
	}
	if build.BuilderID != commander.CommanderID {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "build not found", nil))
		return
	}

	payload := types.PlayerBuildEntry{
		BuildID:    build.ID,
		ShipID:     build.ShipID,
		ShipName:   build.Ship.Name,
		PoolID:     build.PoolID,
		FinishesAt: build.FinishesAt.UTC().Format(time.RFC3339),
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreatePlayerBuild godoc
// @Summary     Create player build
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       payload  body  types.PlayerBuildCreateRequest  true  "Build create"
// @Success     200  {object}  PlayerBuildEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds [post]
func (handler *PlayerHandler) CreatePlayerBuild(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.PlayerBuildCreateRequest
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
	if err := orm.ValidateShipID(req.ShipID); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid ship_id", nil))
		return
	}

	finishAt, err := time.Parse(time.RFC3339, req.FinishesAt)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid finishes_at", nil))
		return
	}

	build := orm.Build{
		BuilderID:  commander.CommanderID,
		ShipID:     req.ShipID,
		PoolID:     req.PoolID,
		FinishesAt: finishAt,
	}
	if err := orm.GormDB.Create(&build).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create build", nil))
		return
	}
	if err := build.Retrieve(true); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load build", nil))
		return
	}

	payload := types.PlayerBuildEntry{
		BuildID:    build.ID,
		ShipID:     build.ShipID,
		ShipName:   build.Ship.Name,
		PoolID:     build.PoolID,
		FinishesAt: build.FinishesAt.UTC().Format(time.RFC3339),
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

	updates := map[string]interface{}{}
	if req.DrawCount1 != nil {
		commander.DrawCount1 = *req.DrawCount1
		updates["draw_count1"] = *req.DrawCount1
	}
	if req.DrawCount10 != nil {
		commander.DrawCount10 = *req.DrawCount10
		updates["draw_count10"] = *req.DrawCount10
	}
	if req.ExchangeCount != nil {
		commander.ExchangeCount = *req.ExchangeCount
		updates["exchange_count"] = *req.ExchangeCount
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(&orm.Commander{}).Where("commander_id = ?", commander.CommanderID).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update counters", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// PlayerSupportRequisition godoc
// @Summary     Get player support requisition counters
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerSupportRequisitionResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/support-requisition [get]
func (handler *PlayerHandler) PlayerSupportRequisition(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	config, err := orm.LoadSupportRequisitionConfig(orm.GormDB)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load support requisition config", nil))
		return
	}

	if commander.EnsureSupportRequisitionMonth(time.Now()) {
		if err := orm.GormDB.Save(&commander).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update support requisition counters", nil))
			return
		}
	}

	payload := types.PlayerSupportRequisitionResponse{
		Month: commander.SupportRequisitionMonth,
		Count: commander.SupportRequisitionCount,
		Cap:   config.MonthlyCap,
	}
	_ = ctx.JSON(response.Success(payload))
}

// ResetPlayerSupportRequisition godoc
// @Summary     Reset player support requisition counters
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerSupportRequisitionResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/support-requisition/reset [post]
func (handler *PlayerHandler) ResetPlayerSupportRequisition(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	config, err := orm.LoadSupportRequisitionConfig(orm.GormDB)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load support requisition config", nil))
		return
	}

	now := time.Now()
	commander.SupportRequisitionMonth = orm.SupportRequisitionMonth(now)
	commander.SupportRequisitionCount = 0
	if err := orm.GormDB.Save(&commander).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to reset support requisition counters", nil))
		return
	}

	payload := types.PlayerSupportRequisitionResponse{
		Month: commander.SupportRequisitionMonth,
		Count: commander.SupportRequisitionCount,
		Cap:   config.MonthlyCap,
	}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerBuild godoc
// @Summary     Update player build
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       build_id  path  int  true  "Build ID"
// @Param       payload   body  types.PlayerBuildUpdateRequest  true  "Build update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds/{build_id} [patch]
func (handler *PlayerHandler) UpdatePlayerBuild(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	buildIDParam := ctx.Params().Get("build_id")
	buildIDValue, err := strconv.ParseUint(buildIDParam, 10, 32)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid build id", nil))
		return
	}

	build, err := orm.GetBuildByID(uint32(buildIDValue))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "build not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load build", nil))
		return
	}
	if build.BuilderID != commander.CommanderID {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "build not found", nil))
		return
	}

	var req types.PlayerBuildUpdateRequest
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
	if req.ShipID != nil {
		if err := orm.ValidateShipID(*req.ShipID); err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid ship_id", nil))
			return
		}
		build.ShipID = *req.ShipID
		updates["ship_id"] = *req.ShipID
	}
	if req.FinishesAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.FinishesAt)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid finishes_at", nil))
			return
		}
		build.FinishesAt = parsed
		updates["finishes_at"] = parsed
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(build).Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update build", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// QuickFinishBuild godoc
// @Summary     Quick finish player build
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       build_id  path  int  true  "Build ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds/{build_id}/quick-finish [patch]
func (handler *PlayerHandler) QuickFinishBuild(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	buildIDParam := ctx.Params().Get("build_id")
	buildIDValue, err := strconv.ParseUint(buildIDParam, 10, 32)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid build id", nil))
		return
	}

	build, err := orm.GetBuildByID(uint32(buildIDValue))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "build not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load build", nil))
		return
	}
	if build.BuilderID != commander.CommanderID {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "build not found", nil))
		return
	}

	build.FinishesAt = time.Now().Add(-24 * time.Hour)
	if err := build.Update(); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to quick finish build", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerBuild godoc
// @Summary     Delete player build
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       build_id  path  int  true  "Build ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/builds/{build_id} [delete]
func (handler *PlayerHandler) DeletePlayerBuild(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	buildID, err := parsePathUint32(ctx.Params().Get("build_id"), "build id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	build, err := orm.GetBuildByID(buildID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "build not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load build", nil))
		return
	}
	if build.BuilderID != commander.CommanderID {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "build not found", nil))
		return
	}

	if err := orm.GormDB.Delete(&orm.Build{}, "id = ? AND builder_id = ?", build.ID, commander.CommanderID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete build", nil))
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
		payload.Mails = append(payload.Mails, buildMailEntry(mail))
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerMail godoc
// @Summary     Get player mail
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       mail_id  path  int  true  "Mail ID"
// @Success     200  {object}  PlayerMailEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/mails/{mail_id} [get]
func (handler *PlayerHandler) PlayerMail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	mailID, err := parsePathUint32(ctx.Params().Get("mail_id"), "mail id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var mail orm.Mail
	if err := orm.GormDB.Preload("Attachments").Where("receiver_id = ?", commander.CommanderID).First(&mail, mailID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "mail not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load mail", nil))
		return
	}

	payload := buildMailEntry(mail)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerMail godoc
// @Summary     Update player mail
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       mail_id  path  int  true  "Mail ID"
// @Param       payload  body  types.PlayerMailUpdateRequest  true  "Mail updates"
// @Success     200  {object}  PlayerMailEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/mails/{mail_id} [patch]
func (handler *PlayerHandler) UpdatePlayerMail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	mailID, err := parsePathUint32(ctx.Params().Get("mail_id"), "mail id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.PlayerMailUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	if req.Read == nil && req.Important == nil && req.Archived == nil && req.AttachmentsCollected == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}

	var mail orm.Mail
	if err := orm.GormDB.Preload("Attachments").Where("receiver_id = ?", commander.CommanderID).First(&mail, mailID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "mail not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load mail", nil))
		return
	}

	updates := map[string]interface{}{}
	if req.Read != nil {
		updates["read"] = *req.Read
	}
	if req.Important != nil {
		updates["is_important"] = *req.Important
	}
	if req.Archived != nil {
		updates["is_archived"] = *req.Archived
	}
	if req.AttachmentsCollected != nil {
		if *req.AttachmentsCollected {
			if !mail.AttachmentsCollected {
				if _, err := mail.CollectAttachments(&commander); err != nil {
					ctx.StatusCode(iris.StatusInternalServerError)
					_ = ctx.JSON(response.Error("internal_error", "failed to collect attachments", nil))
					return
				}
			}
		} else {
			updates["attachments_collected"] = false
		}
	}

	if len(updates) > 0 {
		if err := orm.GormDB.Model(&mail).Updates(updates).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update mail", nil))
			return
		}
	}

	if err := orm.GormDB.Preload("Attachments").Where("receiver_id = ?", commander.CommanderID).First(&mail, mailID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load mail", nil))
		return
	}

	payload := buildMailEntry(mail)
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerMail godoc
// @Summary     Remove player mail
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       mail_id  path  int  true  "Mail ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/mails/{mail_id} [delete]
func (handler *PlayerHandler) DeletePlayerMail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	mailID, err := parsePathUint32(ctx.Params().Get("mail_id"), "mail id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Where("receiver_id = ? AND id = ?", commander.CommanderID, mailID).Delete(&orm.Mail{})
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete mail", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "mail not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
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
		payload.Fleets = append(payload.Fleets, buildFleetEntry(fleet))
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerFleet godoc
// @Summary     Get player fleet
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       fleet_id  path  int  true  "Fleet ID"
// @Success     200  {object}  PlayerFleetEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/fleets/{fleet_id} [get]
func (handler *PlayerHandler) PlayerFleet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	fleetID, err := parsePathUint32(ctx.Params().Get("fleet_id"), "fleet id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	fleet, ok := commander.FleetsMap[fleetID]
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "fleet not found", nil))
		return
	}

	payload := buildFleetEntry(*fleet)
	_ = ctx.JSON(response.Success(payload))
}

// CreatePlayerFleet godoc
// @Summary     Create player fleet
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       payload  body  types.PlayerFleetCreateRequest  true  "Fleet create"
// @Success     200  {object}  PlayerFleetEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/fleets [post]
func (handler *PlayerHandler) CreatePlayerFleet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var req types.PlayerFleetCreateRequest
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

	name := strings.TrimSpace(req.Name)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	if commander.FleetsMap[req.GameID] != nil {
		ctx.StatusCode(iris.StatusConflict)
		_ = ctx.JSON(response.Error("conflict", "fleet already exists", nil))
		return
	}

	if err := orm.CreateFleet(&commander, req.GameID, name, req.ShipIDs); err != nil {
		if errors.Is(err, orm.ErrInvalidShipID) {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "invalid ship_id", nil))
			return
		}
		if errors.Is(err, orm.ErrShipBusy) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "ship is busy", nil))
			return
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "fleet already exists", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create fleet", nil))
		return
	}

	fleet := commander.FleetsMap[req.GameID]
	if fleet == nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load fleet", nil))
		return
	}

	payload := buildFleetEntry(*fleet)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerFleet godoc
// @Summary     Update player fleet
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       fleet_id  path  int  true  "Fleet ID"
// @Param       payload   body  types.PlayerFleetUpdateRequest  true  "Fleet updates"
// @Success     200  {object}  PlayerFleetEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/fleets/{fleet_id} [patch]
func (handler *PlayerHandler) UpdatePlayerFleet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	fleetID, err := parsePathUint32(ctx.Params().Get("fleet_id"), "fleet id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	fleet, ok := commander.FleetsMap[fleetID]
	if !ok {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "fleet not found", nil))
		return
	}

	var req types.PlayerFleetUpdateRequest
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
	if req.Name == nil && req.ShipIDs == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
			return
		}
		if err := fleet.RenameFleet(name); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update fleet", nil))
			return
		}
		fleet.Name = name
	}

	if req.ShipIDs != nil {
		if err := fleet.UpdateShipList(&commander, *req.ShipIDs); err != nil {
			if errors.Is(err, orm.ErrInvalidShipID) {
				ctx.StatusCode(iris.StatusBadRequest)
				_ = ctx.JSON(response.Error("bad_request", "invalid ship_id", nil))
				return
			}
			if errors.Is(err, orm.ErrShipBusy) {
				ctx.StatusCode(iris.StatusConflict)
				_ = ctx.JSON(response.Error("conflict", "ship is busy", nil))
				return
			}
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update fleet", nil))
			return
		}
	}

	payload := buildFleetEntry(*fleet)
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerFleet godoc
// @Summary     Remove player fleet
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       fleet_id  path  int  true  "Fleet ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/fleets/{fleet_id} [delete]
func (handler *PlayerHandler) DeletePlayerFleet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	fleetID, err := parsePathUint32(ctx.Params().Get("fleet_id"), "fleet id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if commander.FleetsMap[fleetID] == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "fleet not found", nil))
		return
	}

	if err := orm.GormDB.Where("commander_id = ? AND game_id = ?", commander.CommanderID, fleetID).
		Delete(&orm.Fleet{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete fleet", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

func buildFleetEntry(fleet orm.Fleet) types.PlayerFleetEntry {
	ships := make([]uint32, 0, len(fleet.ShipList))
	for _, shipID := range fleet.ShipList {
		ships = append(ships, uint32(shipID))
	}
	return types.PlayerFleetEntry{
		FleetID: fleet.GameID,
		Name:    fleet.Name,
		Ships:   ships,
	}
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

	if len(commander.OwnedSkins) == 0 {
		_ = ctx.JSON(response.Success(types.PlayerSkinResponse{Skins: []types.PlayerSkinEntry{}}))
		return
	}

	skinIDs := make([]uint32, 0, len(commander.OwnedSkins))
	for _, owned := range commander.OwnedSkins {
		skinIDs = append(skinIDs, owned.SkinID)
	}

	var skins []orm.Skin
	if err := orm.GormDB.Where("id IN ?", skinIDs).Find(&skins).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load skins", nil))
		return
	}

	skinNames := make(map[uint32]string, len(skins))
	for _, skin := range skins {
		skinNames[skin.ID] = skin.Name
	}

	payload := types.PlayerSkinResponse{Skins: make([]types.PlayerSkinEntry, 0, len(commander.OwnedSkins))}
	for _, owned := range commander.OwnedSkins {
		payload.Skins = append(payload.Skins, types.PlayerSkinEntry{
			SkinID:    owned.SkinID,
			Name:      skinNames[owned.SkinID],
			ExpiresAt: owned.ExpiresAt,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerSkin godoc
// @Summary     Get player skin
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       skin_id  path  int  true  "Skin ID"
// @Success     200  {object}  PlayerSkinEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/skins/{skin_id} [get]
func (handler *PlayerHandler) PlayerSkin(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	skinID, err := parsePathUint32(ctx.Params().Get("skin_id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	owned := commander.OwnedSkinsMap[skinID]
	if owned == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "skin not owned", nil))
		return
	}

	var skin orm.Skin
	if err := orm.GormDB.Select("name").First(&skin, skinID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "skin not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load skin", nil))
		return
	}

	payload := types.PlayerSkinEntry{
		SkinID:    owned.SkinID,
		Name:      skin.Name,
		ExpiresAt: owned.ExpiresAt,
	}

	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerSkin godoc
// @Summary     Update player skin
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       skin_id  path  int  true  "Skin ID"
// @Param       payload  body  types.PlayerSkinUpdateRequest  true  "Skin update"
// @Success     200  {object}  PlayerSkinEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/skins/{skin_id} [patch]
func (handler *PlayerHandler) UpdatePlayerSkin(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	skinID, err := parsePathUint32(ctx.Params().Get("skin_id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	owned := commander.OwnedSkinsMap[skinID]
	if owned == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "skin not owned", nil))
		return
	}

	var req types.PlayerSkinUpdateRequest
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

	parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, req.ExpiresAt)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "expires_at must be RFC3339", nil))
			return
		}
	}
	expiresAt := parsed

	if err := orm.GormDB.Model(&orm.OwnedSkin{}).
		Where("commander_id = ? AND skin_id = ?", commander.CommanderID, skinID).
		Update("expires_at", &expiresAt).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update skin", nil))
		return
	}
	owned.ExpiresAt = &expiresAt

	var skin orm.Skin
	if err := orm.GormDB.Select("name").First(&skin, skinID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "skin not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load skin", nil))
		return
	}

	payload := types.PlayerSkinEntry{
		SkinID:    owned.SkinID,
		Name:      skin.Name,
		ExpiresAt: owned.ExpiresAt,
	}

	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerSkin godoc
// @Summary     Remove player skin
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       skin_id  path  int  true  "Skin ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/skins/{skin_id} [delete]
func (handler *PlayerHandler) DeletePlayerSkin(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	skinID, err := parsePathUint32(ctx.Params().Get("skin_id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if commander.OwnedSkinsMap[skinID] == nil {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "skin not owned", nil))
		return
	}

	if err := orm.GormDB.Where("commander_id = ? AND skin_id = ?", commander.CommanderID, skinID).Delete(&orm.OwnedSkin{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete skin", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
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

// PlayerBuff godoc
// @Summary     Get player buff
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       buff_id  path  int  true  "Buff ID"
// @Success     200  {object}  PlayerBuffEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/buffs/{buff_id} [get]
func (handler *PlayerHandler) PlayerBuff(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	buffID, err := parsePathUint32(ctx.Params().Get("buff_id"), "buff id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var buff orm.CommanderBuff
	if err := orm.GormDB.Where("commander_id = ? AND buff_id = ?", commander.CommanderID, buffID).First(&buff).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "buff not owned", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load buff", nil))
		return
	}

	payload := types.PlayerBuffEntry{
		BuffID:    buff.BuffID,
		ExpiresAt: buff.ExpiresAt.UTC().Format(time.RFC3339),
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

// PlayerLikes godoc
// @Summary     Get player likes
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerLikesResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/likes [get]
func (handler *PlayerHandler) PlayerLikes(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var likes []orm.Like
	if err := orm.GormDB.Select("group_id").
		Where("liker_id = ?", commander.CommanderID).
		Order("group_id asc").
		Find(&likes).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load likes", nil))
		return
	}
	groupIDs := make([]uint32, 0, len(likes))
	for _, like := range likes {
		groupIDs = append(groupIDs, like.GroupID)
	}
	payload := types.PlayerLikesResponse{GroupIDs: groupIDs}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerLike godoc
// @Summary     Add player like
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerLikeRequest  true  "Like request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/likes [post]
func (handler *PlayerHandler) AddPlayerLike(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerLikeRequest
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
	like := orm.Like{GroupID: req.GroupID, LikerID: commander.CommanderID}
	if err := orm.GormDB.Create(&like).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			_ = ctx.JSON(response.Success(nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to add like", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerLike godoc
// @Summary     Remove player like
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       group_id  path  int  true  "Group ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/likes/{group_id} [delete]
func (handler *PlayerHandler) DeletePlayerLike(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	groupID, err := parsePathUint32(ctx.Params().Get("group_id"), "group id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.Where("liker_id = ? AND group_id = ?", commander.CommanderID, groupID).Delete(&orm.Like{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete like", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerRandomFlagShips godoc
// @Summary     Get player random flagships
// @Tags        Players
// @Produce     json
// @Param       id       path  int  true   "Player ID"
// @Param       ship_id  query  int  false  "Filter by ship ID"
// @Success     200  {object}  PlayerRandomFlagShipListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/random-flagships [get]
func (handler *PlayerHandler) PlayerRandomFlagShips(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	shipIDParam := strings.TrimSpace(ctx.URLParam("ship_id"))
	var shipID *uint32
	if shipIDParam != "" {
		parsed, err := parsePathUint32(shipIDParam, "ship id")
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
			return
		}
		shipID = &parsed
	}

	var entries []orm.RandomFlagShip
	query := orm.GormDB.Where("commander_id = ?", commander.CommanderID)
	if shipID != nil {
		query = query.Where("ship_id = ?", *shipID)
	}
	if err := query.Order("ship_id asc").Order("phantom_id asc").Find(&entries).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load random flagships", nil))
		return
	}

	payload := types.PlayerRandomFlagShipListResponse{Entries: make([]types.PlayerRandomFlagShipEntry, 0, len(entries))}
	for _, entry := range entries {
		payload.Entries = append(payload.Entries, types.PlayerRandomFlagShipEntry{
			ShipID:    entry.ShipID,
			PhantomID: entry.PhantomID,
			Enabled:   entry.Enabled,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

// UpsertPlayerRandomFlagShip godoc
// @Summary     Upsert player random flagship
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       payload  body  types.PlayerRandomFlagShipUpsertRequest  true  "Random flagship upsert"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/random-flagships [post]
func (handler *PlayerHandler) UpsertPlayerRandomFlagShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerRandomFlagShipUpsertRequest
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
	if err := orm.ValidateShipID(req.ShipID); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid ship_id", nil))
		return
	}

	entry := orm.RandomFlagShip{
		CommanderID: commander.CommanderID,
		ShipID:      req.ShipID,
		PhantomID:   req.PhantomID,
		Enabled:     req.Enabled,
	}
	if err := orm.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "commander_id"}, {Name: "ship_id"}, {Name: "phantom_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled"}),
	}).Create(&entry).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to upsert random flagship", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerRandomFlagShip godoc
// @Summary     Remove player random flagship
// @Tags        Players
// @Produce     json
// @Param       id          path  int  true  "Player ID"
// @Param       ship_id     path  int  true  "Ship ID"
// @Param       phantom_id  path  int  true  "Phantom ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/random-flagships/{ship_id}/{phantom_id} [delete]
func (handler *PlayerHandler) DeletePlayerRandomFlagShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	shipID, err := parsePathUint32(ctx.Params().Get("ship_id"), "ship id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	phantomID, err := parsePathUint32(ctx.Params().Get("phantom_id"), "phantom id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := orm.GormDB.Where("commander_id = ? AND ship_id = ? AND phantom_id = ?", commander.CommanderID, shipID, phantomID).
		Delete(&orm.RandomFlagShip{}).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete random flagship", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerRandomFlagShip godoc
// @Summary     Get player random flagship toggle
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerRandomFlagShipResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/random-flagship [get]
func (handler *PlayerHandler) PlayerRandomFlagShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	payload := types.PlayerRandomFlagShipResponse{Enabled: commander.RandomFlagShipEnabled}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerRandomFlagShip godoc
// @Summary     Update player random flagship toggle
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerRandomFlagShipRequest  true  "Random flagship request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/random-flagship [patch]
func (handler *PlayerHandler) UpdatePlayerRandomFlagShip(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerRandomFlagShipRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if err := orm.UpdateCommanderRandomFlagShipEnabled(orm.GormDB, commander.CommanderID, req.Enabled); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update random flagship", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerRandomFlagShipMode godoc
// @Summary     Get player random flagship mode
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerRandomFlagShipModeResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/random-flagship-mode [get]
func (handler *PlayerHandler) PlayerRandomFlagShipMode(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	payload := types.PlayerRandomFlagShipModeResponse{Mode: commander.RandomShipMode}
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerRandomFlagShipMode godoc
// @Summary     Update player random flagship mode
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Param       payload  body  types.PlayerRandomFlagShipModeRequest  true  "Random flagship mode request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/random-flagship-mode [patch]
func (handler *PlayerHandler) UpdatePlayerRandomFlagShipMode(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.PlayerRandomFlagShipModeRequest
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
	if err := orm.UpdateCommanderRandomShipMode(orm.GormDB, commander.CommanderID, req.Mode); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update random flagship mode", nil))
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
	if err := orm.GormDB.Model(&orm.Commander{}).Where("commander_id = ?", commander.CommanderID).Updates(updates).Error; err != nil {
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

// UpsertPlayerStory godoc
// @Summary     Add player story progress
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       story_id  path  int  true  "Story ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/stories/{story_id} [put]
func (handler *PlayerHandler) UpsertPlayerStory(ctx iris.Context) {
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
	if err := orm.AddCommanderStory(orm.GormDB, commander.CommanderID, storyID); err != nil {
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
			parsed, err = time.Parse(time.RFC3339Nano, *req.ExpiresAt)
			if err != nil {
				ctx.StatusCode(iris.StatusBadRequest)
				_ = ctx.JSON(response.Error("bad_request", "expires_at must be RFC3339", nil))
				return
			}
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

// UpdatePlayerAttire godoc
// @Summary     Update player attire
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id         path  int  true  "Player ID"
// @Param       type       path  int  true  "Attire type"
// @Param       attire_id  path  int  true  "Attire ID"
// @Param       payload    body  types.PlayerAttireUpdateRequest  true  "Attire update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/attires/{type}/{attire_id} [patch]
func (handler *PlayerHandler) UpdatePlayerAttire(ctx iris.Context) {
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
	var existing orm.CommanderAttire
	if err := orm.GormDB.Where("commander_id = ? AND type = ? AND attire_id = ?", commander.CommanderID, attireType, attireID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "attire not owned", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load attire", nil))
		return
	}
	var req types.PlayerAttireUpdateRequest
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
	if req.ExpiresAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339Nano, *req.ExpiresAt)
			if err != nil {
				ctx.StatusCode(iris.StatusBadRequest)
				_ = ctx.JSON(response.Error("bad_request", "expires_at must be RFC3339", nil))
				return
			}
		}
		expiresAt := parsed
		updates["expires_at"] = &expiresAt
	}
	if req.IsNew != nil {
		updates["is_new"] = *req.IsNew
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(&orm.CommanderAttire{}).
		Where("commander_id = ? AND type = ? AND attire_id = ?", commander.CommanderID, attireType, attireID).
		Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update attire", nil))
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
	if err := orm.GormDB.Model(&orm.Commander{}).Where("commander_id = ?", commander.CommanderID).Updates(updates).Error; err != nil {
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

// UpdatePlayerLivingAreaCoverState godoc
// @Summary     Update player living area cover
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       cover_id  path  int  true  "Cover ID"
// @Param       payload   body  types.PlayerLivingAreaCoverUpdateRequest  true  "Cover update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/livingarea-covers/{cover_id} [patch]
func (handler *PlayerHandler) UpdatePlayerLivingAreaCoverState(ctx iris.Context) {
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
	var existing orm.CommanderLivingAreaCover
	if err := orm.GormDB.Where("commander_id = ? AND cover_id = ?", commander.CommanderID, coverID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "cover not owned", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load cover", nil))
		return
	}
	var req types.PlayerLivingAreaCoverUpdateRequest
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
	if req.IsNew != nil {
		updates["is_new"] = *req.IsNew
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if err := orm.GormDB.Model(&orm.CommanderLivingAreaCover{}).
		Where("commander_id = ? AND cover_id = ?", commander.CommanderID, coverID).
		Updates(updates).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update cover", nil))
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
	if err := orm.GormDB.Model(&orm.Commander{}).Where("commander_id = ?", commander.CommanderID).Update("living_area_cover_id", req.CoverID).Error; err != nil {
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

// DeletePlayerShoppingStreet godoc
// @Summary     Clear player shopping street
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street [delete]
func (handler *PlayerHandler) DeletePlayerShoppingStreet(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	if err := orm.GormDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("commander_id = ?", commander.CommanderID).Delete(&orm.ShoppingStreetGood{}).Error; err != nil {
			return err
		}
		return tx.Where("commander_id = ?", commander.CommanderID).Delete(&orm.ShoppingStreetState{}).Error
	}); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to clear shopping street", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// PlayerShoppingStreetGoods godoc
// @Summary     Get player shopping street goods
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerShoppingStreetGoodsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street/goods [get]
func (handler *PlayerHandler) PlayerShoppingStreetGoods(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	_, goods, err := shopstreet.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street goods", nil))
		return
	}
	payload := types.ShoppingStreetGoodsResponse{Goods: make([]types.ShoppingStreetGood, 0, len(goods))}
	for _, good := range goods {
		payload.Goods = append(payload.Goods, types.ShoppingStreetGood{
			GoodsID:  good.GoodsID,
			Discount: good.Discount,
			BuyCount: good.BuyCount,
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

// PlayerShoppingStreetGood godoc
// @Summary     Get player shopping street good
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       goods_id  path  int  true  "Goods ID"
// @Success     200  {object}  PlayerShoppingStreetGoodResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street/goods/{goods_id} [get]
func (handler *PlayerHandler) PlayerShoppingStreetGood(ctx iris.Context) {
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
	var good orm.ShoppingStreetGood
	if err := orm.GormDB.Where("commander_id = ? AND goods_id = ?", commander.CommanderID, goodsID).First(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "goods not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load goods", nil))
		return
	}
	payload := types.ShoppingStreetGood{
		GoodsID:  good.GoodsID,
		Discount: good.Discount,
		BuyCount: good.BuyCount,
	}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerShoppingStreetGood godoc
// @Summary     Add player shopping street good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id      path  int  true  "Player ID"
// @Param       body    body  types.ShoppingStreetGoodCreateRequest  true  "Shopping street good"
// @Success     200  {object}  PlayerShoppingStreetResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/shopping-street/goods [post]
func (handler *PlayerHandler) AddPlayerShoppingStreetGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.ShoppingStreetGoodCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if req.Discount < 1 || req.Discount > 100 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "discount must be between 1 and 100", nil))
		return
	}
	_, invalid, err := shopstreet.ResolveOffers([]uint32{req.GoodsID})
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
	state, _, err := shopstreet.EnsureState(commander.CommanderID, time.Now())
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street", nil))
		return
	}
	good := orm.ShoppingStreetGood{
		CommanderID: commander.CommanderID,
		GoodsID:     req.GoodsID,
		Discount:    req.Discount,
		BuyCount:    req.BuyCount,
	}
	if err := orm.GormDB.Create(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "goods already exists", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create goods", nil))
		return
	}
	goods, err := shopstreet.LoadGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load shopping street goods", nil))
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

// PlayerPunishments godoc
// @Summary     Get player punishments
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  PlayerPunishmentsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/punishments [get]
func (handler *PlayerHandler) PlayerPunishments(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var commander orm.Commander
	if err := orm.GormDB.Select("commander_id").First(&commander, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var punishments []orm.Punishment
	if err := orm.GormDB.Where("punished_id = ?", commanderID).Order("id desc").Find(&punishments).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load punishments", nil))
		return
	}

	payload := types.PlayerPunishmentsResponse{Punishments: make([]types.PlayerPunishmentEntry, 0, len(punishments))}
	for _, punishment := range punishments {
		payload.Punishments = append(payload.Punishments, buildPunishmentEntry(punishment))
	}

	_ = ctx.JSON(response.Success(payload))
}

// PlayerPunishment godoc
// @Summary     Get player punishment
// @Tags        Players
// @Produce     json
// @Param       id              path  int  true  "Player ID"
// @Param       punishment_id   path  int  true  "Punishment ID"
// @Success     200  {object}  PlayerPunishmentEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/punishments/{punishment_id} [get]
func (handler *PlayerHandler) PlayerPunishment(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	punishmentID, err := parsePathUint32(ctx.Params().Get("punishment_id"), "punishment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var commander orm.Commander
	if err := orm.GormDB.Select("commander_id").First(&commander, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}

	var punishment orm.Punishment
	if err := orm.GormDB.Where("punished_id = ?", commanderID).First(&punishment, punishmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "punishment not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load punishment", nil))
		return
	}

	payload := buildPunishmentEntry(punishment)
	_ = ctx.JSON(response.Success(payload))
}

// CreatePlayerPunishment godoc
// @Summary     Create punishment for player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int                    true  "Player ID"
// @Param       payload  body  types.BanPlayerRequest  true  "Punishment request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/punishments [post]
func (handler *PlayerHandler) CreatePlayerPunishment(ctx iris.Context) {
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
		_ = ctx.JSON(response.Error("bad_request", "punishment requires duration_sec, lift_timestamp, or permanent=true", nil))
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
		_ = ctx.JSON(response.Error("internal_error", "failed to create punishment", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdatePlayerPunishment godoc
// @Summary     Update player punishment
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id              path  int                                  true  "Player ID"
// @Param       punishment_id   path  int                                  true  "Punishment ID"
// @Param       payload         body  types.PlayerPunishmentUpdateRequest  true  "Punishment update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/punishments/{punishment_id} [patch]
func (handler *PlayerHandler) UpdatePlayerPunishment(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	punishmentID, err := parsePathUint32(ctx.Params().Get("punishment_id"), "punishment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.PlayerPunishmentUpdateRequest
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
	if req.Permanent == nil && req.LiftTimestamp == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	if req.Permanent != nil && *req.Permanent && req.LiftTimestamp != nil && strings.TrimSpace(*req.LiftTimestamp) != "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "permanent cannot be combined with lift_timestamp", nil))
		return
	}

	var commander orm.Commander
	if err := orm.GormDB.Select("commander_id").First(&commander, commanderID).Error; err != nil {
		writeCommanderError(ctx, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Permanent != nil {
		updates["is_permanent"] = *req.Permanent
		if *req.Permanent {
			updates["lift_timestamp"] = nil
		}
	}
	if req.LiftTimestamp != nil && (req.Permanent == nil || !*req.Permanent) {
		if strings.TrimSpace(*req.LiftTimestamp) == "" {
			updates["lift_timestamp"] = nil
		} else {
			parsed, err := time.Parse(time.RFC3339, *req.LiftTimestamp)
			if err != nil {
				ctx.StatusCode(iris.StatusBadRequest)
				_ = ctx.JSON(response.Error("bad_request", "lift_timestamp must be RFC3339", nil))
				return
			}
			updates["lift_timestamp"] = parsed
		}
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}

	result := orm.GormDB.Model(&orm.Punishment{}).Where("punished_id = ? AND id = ?", commanderID, punishmentID).Updates(updates)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update punishment", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "punishment not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeletePlayerPunishment godoc
// @Summary     Delete player punishment
// @Tags        Players
// @Produce     json
// @Param       id              path  int  true  "Player ID"
// @Param       punishment_id   path  int  true  "Punishment ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/punishments/{punishment_id} [delete]
func (handler *PlayerHandler) DeletePlayerPunishment(ctx iris.Context) {
	commanderID, err := parseCommanderID(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	punishmentID, err := parsePathUint32(ctx.Params().Get("punishment_id"), "punishment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Where("punished_id = ? AND id = ?", commanderID, punishmentID).Delete(&orm.Punishment{})
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete punishment", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "punishment not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// BanPlayer godoc
// @Summary     Ban player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int                    true  "Player ID"
// @Param       payload  body  types.BanPlayerRequest  true  "Ban request"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ban [post]
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

// UnbanPlayer godoc
// @Summary     Unban player
// @Tags        Players
// @Produce     json
// @Param       id   path  int  true  "Player ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/ban [delete]
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

// KickPlayer godoc
// @Summary     Kick player
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int                     true   "Player ID"
// @Param       payload  body  types.KickPlayerRequest  false  "Kick request"
// @Success     200  {object}  KickPlayerResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/kick [post]
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

	if err := commander.GiveSkinWithExpiry(req.SkinID, expiresAt); err != nil {
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

// UpdatePlayerBuff godoc
// @Summary     Update player buff
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Player ID"
// @Param       buff_id  path  int  true  "Buff ID"
// @Param       payload  body  types.PlayerBuffUpdateRequest  true  "Buff update"
// @Success     200  {object}  PlayerBuffEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/buffs/{buff_id} [patch]
func (handler *PlayerHandler) UpdatePlayerBuff(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	buffID, err := parsePathUint32(ctx.Params().Get("buff_id"), "buff id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.PlayerBuffUpdateRequest
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

	var buff orm.CommanderBuff
	if err := orm.GormDB.Where("commander_id = ? AND buff_id = ?", commander.CommanderID, buffID).First(&buff).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "buff not owned", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load buff", nil))
		return
	}

	if err := orm.GormDB.Model(&orm.CommanderBuff{}).
		Where("commander_id = ? AND buff_id = ?", commander.CommanderID, buffID).
		Update("expires_at", expiresAt.UTC()).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update buff", nil))
		return
	}

	payload := types.PlayerBuffEntry{
		BuffID:    buffID,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	}

	_ = ctx.JSON(response.Success(payload))
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
	commander.RebuildOwnedEquipmentMap()
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

func buildMailEntry(mail orm.Mail) types.PlayerMailEntry {
	attachments := make([]types.PlayerMailAttachment, 0, len(mail.Attachments))
	for _, attachment := range mail.Attachments {
		attachments = append(attachments, types.PlayerMailAttachment{
			Type:     attachment.Type,
			ItemID:   attachment.ItemID,
			Quantity: attachment.Quantity,
		})
	}

	return types.PlayerMailEntry{
		MailID:               mail.ID,
		Title:                mail.Title,
		Body:                 mail.Body,
		Read:                 mail.Read,
		Date:                 mail.Date.UTC().Format(time.RFC3339),
		Important:            mail.IsImportant,
		Archived:             mail.IsArchived,
		AttachmentsCollected: mail.AttachmentsCollected,
		Sender:               mail.CustomSender,
		Attachments:          attachments,
	}
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

func buildPunishmentEntry(punishment orm.Punishment) types.PlayerPunishmentEntry {
	liftTimestamp := ""
	if punishment.LiftTimestamp != nil {
		liftTimestamp = punishment.LiftTimestamp.UTC().Format(time.RFC3339)
	}

	return types.PlayerPunishmentEntry{
		PunishmentID:  punishment.ID,
		Permanent:     punishment.IsPermanent,
		LiftTimestamp: liftTimestamp,
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

// DeletePlayerArenaShop godoc
// @Summary     Clear player arena shop state
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerArenaShopDeleteResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/arena-shop [delete]
func (handler *PlayerHandler) DeletePlayerArenaShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	result := orm.GormDB.Delete(&orm.ArenaShopState{}, "commander_id = ?", commander.CommanderID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete arena shop state", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "arena shop state not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
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

// PlayerMedalShopGoods godoc
// @Summary     Get player medal shop goods
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerMedalShopGoodsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/medal-shop/goods [get]
func (handler *PlayerHandler) PlayerMedalShopGoods(ctx iris.Context) {
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
	_, goods, err := medalshop.RefreshIfNeeded(commander.CommanderID, time.Now(), config)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop goods", nil))
		return
	}
	payload := types.MedalShopGoodsResponse{Goods: make([]types.MedalShopGood, 0, len(goods))}
	for _, good := range goods {
		payload.Goods = append(payload.Goods, types.MedalShopGood{
			Index:   good.Index,
			GoodsID: good.GoodsID,
			Count:   good.Count,
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerMedalShopGood godoc
// @Summary     Add player medal shop good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id    path  int  true  "Player ID"
// @Param       body  body  types.MedalShopGoodCreateRequest  true  "Medal shop good"
// @Success     200  {object}  PlayerMedalShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/medal-shop/goods [post]
func (handler *PlayerHandler) AddPlayerMedalShopGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.MedalShopGoodCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if req.Index == 0 || req.GoodsID == 0 || req.Count == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "index, goods_id, and count are required", nil))
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
	good := orm.MedalShopGood{
		CommanderID: commander.CommanderID,
		Index:       req.Index,
		GoodsID:     req.GoodsID,
		Count:       req.Count,
	}
	if err := orm.GormDB.Create(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "goods already exists", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create goods", nil))
		return
	}
	goods, err := medalshop.LoadGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop goods", nil))
		return
	}
	payload := buildMedalShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerMedalShopGood godoc
// @Summary     Update player medal shop good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id     path  int  true  "Player ID"
// @Param       index  path  int  true  "Goods index"
// @Param       body   body  types.MedalShopGoodPatchRequest  true  "Good updates"
// @Success     200  {object}  PlayerMedalShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/medal-shop/goods/{index} [patch]
func (handler *PlayerHandler) UpdatePlayerMedalShopGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	index, err := parsePathUint32(ctx.Params().Get("index"), "index")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.MedalShopGoodPatchRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.GoodsID != nil {
		if *req.GoodsID == 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "goods_id must be > 0", nil))
			return
		}
		updates["goods_id"] = *req.GoodsID
	}
	if req.Count != nil {
		if *req.Count == 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "count must be > 0", nil))
			return
		}
		updates["count"] = *req.Count
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	result := orm.GormDB.Model(&orm.MedalShopGood{}).
		Where("commander_id = ? AND index = ?", commander.CommanderID, index).
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
	var state orm.MedalShopState
	if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "medal shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop state", nil))
		return
	}
	goods, err := medalshop.LoadGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop goods", nil))
		return
	}
	payload := buildMedalShopResponse(&state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerMedalShopGood godoc
// @Summary     Delete player medal shop good
// @Tags        Players
// @Produce     json
// @Param       id     path  int  true  "Player ID"
// @Param       index  path  int  true  "Goods index"
// @Success     200  {object}  PlayerMedalShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/medal-shop/goods/{index} [delete]
func (handler *PlayerHandler) DeletePlayerMedalShopGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	index, err := parsePathUint32(ctx.Params().Get("index"), "index")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	result := orm.GormDB.Where("commander_id = ? AND index = ?", commander.CommanderID, index).Delete(&orm.MedalShopGood{})
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
	var state orm.MedalShopState
	if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "medal shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop state", nil))
		return
	}
	goods, err := medalshop.LoadGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load medal shop goods", nil))
		return
	}
	payload := buildMedalShopResponse(&state, goods)
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

// PlayerGuildShop godoc
// @Summary     Get player guild shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerGuildShopResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guild-shop [get]
func (handler *PlayerHandler) PlayerGuildShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	state, err := loadGuildShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "guild shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop state", nil))
		return
	}
	goods, err := loadGuildShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop goods", nil))
		return
	}
	payload := buildGuildShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// RefreshPlayerGuildShop godoc
// @Summary     Refresh player guild shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerGuildShopResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guild-shop/refresh [post]
func (handler *PlayerHandler) RefreshPlayerGuildShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	state, err := loadGuildShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "guild shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop state", nil))
		return
	}
	goods, err := loadGuildShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop goods", nil))
		return
	}
	payload := buildGuildShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerGuildShop godoc
// @Summary     Update player guild shop state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id    path  int  true  "Player ID"
// @Param       body  body  types.GuildShopUpdateRequest  true  "State updates"
// @Success     200  {object}  PlayerGuildShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guild-shop [put]
func (handler *PlayerHandler) UpdatePlayerGuildShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.GuildShopUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.RefreshCount != nil {
		updates["refresh_count"] = *req.RefreshCount
	}
	if req.NextRefreshTime != nil {
		updates["next_refresh_time"] = *req.NextRefreshTime
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	var state orm.GuildShopState
	if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop state", nil))
			return
		}
		state = orm.GuildShopState{CommanderID: commander.CommanderID}
		if req.RefreshCount != nil {
			state.RefreshCount = *req.RefreshCount
		}
		if req.NextRefreshTime != nil {
			state.NextRefreshTime = *req.NextRefreshTime
		}
		if err := orm.GormDB.Create(&state).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to create guild shop state", nil))
			return
		}
	} else {
		if err := orm.GormDB.Model(&orm.GuildShopState{}).
			Where("commander_id = ?", commander.CommanderID).
			Updates(updates).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update guild shop state", nil))
			return
		}
		if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to reload guild shop state", nil))
			return
		}
	}
	goods, err := loadGuildShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop goods", nil))
		return
	}
	payload := buildGuildShopResponse(&state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// PlayerGuildShopGoods godoc
// @Summary     Get player guild shop goods
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerGuildShopGoodsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guild-shop/goods [get]
func (handler *PlayerHandler) PlayerGuildShopGoods(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	goods, err := loadGuildShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop goods", nil))
		return
	}
	payload := types.GuildShopGoodsResponse{Goods: make([]types.GuildShopGood, 0, len(goods))}
	for _, good := range goods {
		payload.Goods = append(payload.Goods, types.GuildShopGood{
			Index:   good.Index,
			GoodsID: good.GoodsID,
			Count:   good.Count,
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerGuildShopGood godoc
// @Summary     Add player guild shop good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id    path  int  true  "Player ID"
// @Param       body  body  types.GuildShopGoodCreateRequest  true  "Guild shop good"
// @Success     200  {object}  PlayerGuildShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guild-shop/goods [post]
func (handler *PlayerHandler) AddPlayerGuildShopGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.GuildShopGoodCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if req.Index == 0 || req.GoodsID == 0 || req.Count == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "index, goods_id, and count are required", nil))
		return
	}
	state, err := loadGuildShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "guild shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop state", nil))
		return
	}
	good := orm.GuildShopGood{
		CommanderID: commander.CommanderID,
		Index:       req.Index,
		GoodsID:     req.GoodsID,
		Count:       req.Count,
	}
	if err := orm.GormDB.Create(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "goods already exists", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create goods", nil))
		return
	}
	goods, err := loadGuildShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop goods", nil))
		return
	}
	payload := buildGuildShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerGuildShopGood godoc
// @Summary     Update player guild shop good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id     path  int  true  "Player ID"
// @Param       index  path  int  true  "Goods index"
// @Param       body   body  types.GuildShopGoodPatchRequest  true  "Good updates"
// @Success     200  {object}  PlayerGuildShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guild-shop/goods/{index} [patch]
func (handler *PlayerHandler) UpdatePlayerGuildShopGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	index, err := parsePathUint32(ctx.Params().Get("index"), "index")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.GuildShopGoodPatchRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.GoodsID != nil {
		if *req.GoodsID == 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "goods_id must be > 0", nil))
			return
		}
		updates["goods_id"] = *req.GoodsID
	}
	if req.Count != nil {
		if *req.Count == 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "count must be > 0", nil))
			return
		}
		updates["count"] = *req.Count
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	result := orm.GormDB.Model(&orm.GuildShopGood{}).
		Where("commander_id = ? AND index = ?", commander.CommanderID, index).
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
	state, err := loadGuildShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "guild shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop state", nil))
		return
	}
	goods, err := loadGuildShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop goods", nil))
		return
	}
	payload := buildGuildShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerGuildShopGood godoc
// @Summary     Delete player guild shop good
// @Tags        Players
// @Produce     json
// @Param       id     path  int  true  "Player ID"
// @Param       index  path  int  true  "Goods index"
// @Success     200  {object}  PlayerGuildShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/guild-shop/goods/{index} [delete]
func (handler *PlayerHandler) DeletePlayerGuildShopGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	index, err := parsePathUint32(ctx.Params().Get("index"), "index")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	result := orm.GormDB.Where("commander_id = ? AND index = ?", commander.CommanderID, index).Delete(&orm.GuildShopGood{})
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
	state, err := loadGuildShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "guild shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop state", nil))
		return
	}
	goods, err := loadGuildShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load guild shop goods", nil))
		return
	}
	payload := buildGuildShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// PlayerMiniGameShop godoc
// @Summary     Get player minigame shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerMiniGameShopResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/minigame-shop [get]
func (handler *PlayerHandler) PlayerMiniGameShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	state, err := loadMiniGameShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "minigame shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop state", nil))
		return
	}
	goods, err := loadMiniGameShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop goods", nil))
		return
	}
	payload := buildMiniGameShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// RefreshPlayerMiniGameShop godoc
// @Summary     Refresh player minigame shop
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerMiniGameShopResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/minigame-shop/refresh [post]
func (handler *PlayerHandler) RefreshPlayerMiniGameShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	state, err := loadMiniGameShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "minigame shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop state", nil))
		return
	}
	goods, err := loadMiniGameShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop goods", nil))
		return
	}
	payload := buildMiniGameShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerMiniGameShop godoc
// @Summary     Update player minigame shop state
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id    path  int  true  "Player ID"
// @Param       body  body  types.MiniGameShopUpdateRequest  true  "State updates"
// @Success     200  {object}  PlayerMiniGameShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/minigame-shop [put]
func (handler *PlayerHandler) UpdatePlayerMiniGameShop(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.MiniGameShopUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.NextRefreshTime != nil {
		updates["next_refresh_time"] = *req.NextRefreshTime
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	var state orm.MiniGameShopState
	if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop state", nil))
			return
		}
		state = orm.MiniGameShopState{CommanderID: commander.CommanderID}
		if req.NextRefreshTime != nil {
			state.NextRefreshTime = *req.NextRefreshTime
		}
		if err := orm.GormDB.Create(&state).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to create minigame shop state", nil))
			return
		}
	} else {
		if err := orm.GormDB.Model(&orm.MiniGameShopState{}).
			Where("commander_id = ?", commander.CommanderID).
			Updates(updates).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update minigame shop state", nil))
			return
		}
		if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).First(&state).Error; err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to reload minigame shop state", nil))
			return
		}
	}
	goods, err := loadMiniGameShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop goods", nil))
		return
	}
	payload := buildMiniGameShopResponse(&state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// PlayerMiniGameShopGoods godoc
// @Summary     Get player minigame shop goods
// @Tags        Players
// @Produce     json
// @Param       id  path  int  true  "Player ID"
// @Success     200  {object}  PlayerMiniGameShopGoodsResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/minigame-shop/goods [get]
func (handler *PlayerHandler) PlayerMiniGameShopGoods(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	goods, err := loadMiniGameShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop goods", nil))
		return
	}
	payload := types.MiniGameShopGoodsResponse{Goods: make([]types.MiniGameShopGood, 0, len(goods))}
	for _, good := range goods {
		payload.Goods = append(payload.Goods, types.MiniGameShopGood{
			GoodsID: good.GoodsID,
			Count:   good.Count,
		})
	}
	_ = ctx.JSON(response.Success(payload))
}

// AddPlayerMiniGameShopGood godoc
// @Summary     Add player minigame shop good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id    path  int  true  "Player ID"
// @Param       body  body  types.MiniGameShopGoodCreateRequest  true  "Minigame shop good"
// @Success     200  {object}  PlayerMiniGameShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     409  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/minigame-shop/goods [post]
func (handler *PlayerHandler) AddPlayerMiniGameShopGood(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.MiniGameShopGoodCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	if req.GoodsID == 0 || req.Count == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "goods_id and count are required", nil))
		return
	}
	state, err := loadMiniGameShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "minigame shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop state", nil))
		return
	}
	good := orm.MiniGameShopGood{
		CommanderID: commander.CommanderID,
		GoodsID:     req.GoodsID,
		Count:       req.Count,
	}
	if err := orm.GormDB.Create(&good).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			ctx.StatusCode(iris.StatusConflict)
			_ = ctx.JSON(response.Error("conflict", "goods already exists", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create goods", nil))
		return
	}
	goods, err := loadMiniGameShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop goods", nil))
		return
	}
	payload := buildMiniGameShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// UpdatePlayerMiniGameShopGood godoc
// @Summary     Update player minigame shop good
// @Tags        Players
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       goods_id  path  int  true  "Goods ID"
// @Param       body      body  types.MiniGameShopGoodPatchRequest  true  "Good updates"
// @Success     200  {object}  PlayerMiniGameShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/minigame-shop/goods/{goods_id} [patch]
func (handler *PlayerHandler) UpdatePlayerMiniGameShopGood(ctx iris.Context) {
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
	var req types.MiniGameShopGoodPatchRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	updates := map[string]interface{}{}
	if req.Count != nil {
		if *req.Count == 0 {
			ctx.StatusCode(iris.StatusBadRequest)
			_ = ctx.JSON(response.Error("bad_request", "count must be > 0", nil))
			return
		}
		updates["count"] = *req.Count
	}
	if len(updates) == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "no updates provided", nil))
		return
	}
	result := orm.GormDB.Model(&orm.MiniGameShopGood{}).
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
	state, err := loadMiniGameShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "minigame shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop state", nil))
		return
	}
	goods, err := loadMiniGameShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop goods", nil))
		return
	}
	payload := buildMiniGameShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

// DeletePlayerMiniGameShopGood godoc
// @Summary     Delete player minigame shop good
// @Tags        Players
// @Produce     json
// @Param       id        path  int  true  "Player ID"
// @Param       goods_id  path  int  true  "Goods ID"
// @Success     200  {object}  PlayerMiniGameShopResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/minigame-shop/goods/{goods_id} [delete]
func (handler *PlayerHandler) DeletePlayerMiniGameShopGood(ctx iris.Context) {
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
	result := orm.GormDB.Where("commander_id = ? AND goods_id = ?", commander.CommanderID, goodsID).Delete(&orm.MiniGameShopGood{})
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
	state, err := loadMiniGameShopState(commander.CommanderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "minigame shop state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop state", nil))
		return
	}
	goods, err := loadMiniGameShopGoods(commander.CommanderID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load minigame shop goods", nil))
		return
	}
	payload := buildMiniGameShopResponse(state, goods)
	_ = ctx.JSON(response.Success(payload))
}

func loadGuildShopState(commanderID uint32) (*orm.GuildShopState, error) {
	var state orm.GuildShopState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func loadGuildShopGoods(commanderID uint32) ([]orm.GuildShopGood, error) {
	var goods []orm.GuildShopGood
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Find(&goods).Error; err != nil {
		return nil, err
	}
	return goods, nil
}

func buildGuildShopResponse(state *orm.GuildShopState, goods []orm.GuildShopGood) types.GuildShopResponse {
	response := types.GuildShopResponse{
		State: types.GuildShopState{
			RefreshCount:    state.RefreshCount,
			NextRefreshTime: state.NextRefreshTime,
		},
		Goods: make([]types.GuildShopGood, 0, len(goods)),
	}
	for _, good := range goods {
		response.Goods = append(response.Goods, types.GuildShopGood{
			Index:   good.Index,
			GoodsID: good.GoodsID,
			Count:   good.Count,
		})
	}
	return response
}

func loadMiniGameShopState(commanderID uint32) (*orm.MiniGameShopState, error) {
	var state orm.MiniGameShopState
	if err := orm.GormDB.Where("commander_id = ?", commanderID).First(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

func loadMiniGameShopGoods(commanderID uint32) ([]orm.MiniGameShopGood, error) {
	var goods []orm.MiniGameShopGood
	if err := orm.GormDB.Where("commander_id = ?", commanderID).Find(&goods).Error; err != nil {
		return nil, err
	}
	return goods, nil
}

func buildMiniGameShopResponse(state *orm.MiniGameShopState, goods []orm.MiniGameShopGood) types.MiniGameShopResponse {
	response := types.MiniGameShopResponse{
		State: types.MiniGameShopState{
			NextRefreshTime: state.NextRefreshTime,
		},
		Goods: make([]types.MiniGameShopGood, 0, len(goods)),
	}
	for _, good := range goods {
		response.Goods = append(response.Goods, types.MiniGameShopGood{
			GoodsID: good.GoodsID,
			Count:   good.Count,
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
