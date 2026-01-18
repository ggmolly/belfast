package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/consts"
	"github.com/ggmolly/belfast/internal/logger"
	"github.com/ggmolly/belfast/internal/orm"
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
	party.Get("/{id:uint}/resources", handler.PlayerResources)
	party.Get("/{id:uint}/ships", handler.PlayerShips)
	party.Get("/{id:uint}/items", handler.PlayerItems)
	party.Get("/{id:uint}/builds", handler.PlayerBuilds)
	party.Get("/{id:uint}/mails", handler.PlayerMails)
	party.Get("/{id:uint}/fleets", handler.PlayerFleets)
	party.Get("/{id:uint}/skins", handler.PlayerSkins)
	party.Post("/{id:uint}/ban", handler.BanPlayer)
	party.Delete("/{id:uint}/ban", handler.UnbanPlayer)
	party.Post("/{id:uint}/kick", handler.KickPlayer)
	party.Put("/{id:uint}/resources", handler.UpdateResources)
	party.Post("/{id:uint}/give-ship", handler.GiveShip)
	party.Post("/{id:uint}/give-item", handler.GiveItem)
	party.Post("/{id:uint}/send-mail", handler.SendMail)
	party.Post("/{id:uint}/give-skin", handler.GiveSkin)
	party.Delete("/{id:uint}", handler.DeletePlayer)
}

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

func (handler *PlayerHandler) PlayerBuilds(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}

	payload := types.PlayerBuildResponse{Builds: make([]types.PlayerBuildEntry, 0, len(commander.Builds))}
	for _, build := range commander.Builds {
		payload.Builds = append(payload.Builds, types.PlayerBuildEntry{
			BuildID:    build.ID,
			ShipID:     build.ShipID,
			ShipName:   build.Ship.Name,
			FinishesAt: build.FinishesAt.UTC().Format(time.RFC3339),
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

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

	_ = ctx.JSON(response.Success(nil))
}

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
		ReceiverID: commander.CommanderID,
		Title:      req.Title,
		Body:       req.Body,
	}
	for _, attachment := range req.Attachments {
		mail.Attachments = append(mail.Attachments, orm.MailAttachment{
			Type:     attachment.Type,
			ItemID:   attachment.ItemID,
			Quantity: attachment.Quantity,
		})
	}

	if err := mail.Create(); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to send mail", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

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
	offset, err := parseQueryInt(ctx.URLParamDefault("offset", "0"))
	if err != nil || offset < 0 {
		return orm.PlayerQueryParams{}, fmt.Errorf("offset must be >= 0")
	}
	limit, err := parseQueryInt(ctx.URLParamDefault("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit < 1 {
		return orm.PlayerQueryParams{}, fmt.Errorf("limit must be >= 1")
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	sort := strings.TrimSpace(ctx.URLParam("sort"))
	if sort != "" && sort != "last_login" {
		return orm.PlayerQueryParams{}, fmt.Errorf("unsupported sort")
	}

	filters := strings.TrimSpace(ctx.URLParam("filter"))
	params := orm.PlayerQueryParams{
		Offset:   offset,
		Limit:    limit,
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
