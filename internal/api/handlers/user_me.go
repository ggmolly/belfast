package handlers

import (
	"sort"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/middleware"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/auth"
	"github.com/ggmolly/belfast/internal/authz"
	"github.com/ggmolly/belfast/internal/orm"
)

type MeHandler struct {
	Validate *validator.Validate
}

func NewMeHandler() *MeHandler {
	return &MeHandler{Validate: validator.New(validator.WithRequiredStructEnabled())}
}

func RegisterMeRoutes(party iris.Party, handler *MeHandler) {
	party.Get("/permissions", handler.Permissions)
	party.Get("/commander", handler.Commander)
	party.Get("/resources", middleware.RequirePermission(authz.PermMeResources, authz.ReadSelf), handler.Resources)
	party.Put("/resources", middleware.RequirePermission(authz.PermMeResources, authz.WriteSelf), handler.UpdateResources)
	party.Post("/give-ship", middleware.RequirePermission(authz.PermMeShips, authz.WriteSelf), handler.GiveShip)
	party.Post("/give-item", middleware.RequirePermission(authz.PermMeItems, authz.WriteSelf), handler.GiveItem)
	party.Post("/give-skin", middleware.RequirePermission(authz.PermMeSkins, authz.WriteSelf), handler.GiveSkin)
}

// MePermissions godoc
// @Summary     Get effective permissions
// @Tags        Me
// @Produce     json
// @Success     200  {object}  MePermissionsResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/me/permissions [get]
func (handler *MeHandler) Permissions(ctx iris.Context) {
	account, ok := middleware.GetAccount(ctx)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
		return
	}
	roles, err := orm.ListAccountRoleNames(account.ID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load roles", nil))
		return
	}
	perms, err := orm.LoadEffectivePermissions(account.ID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load permissions", nil))
		return
	}
	entries := make([]types.PermissionPolicyEntry, 0, len(perms))
	for key, cap := range perms {
		entries = append(entries, types.PermissionPolicyEntry{Key: key, ReadSelf: cap.ReadSelf, ReadAny: cap.ReadAny, WriteSelf: cap.WriteSelf, WriteAny: cap.WriteAny})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	_ = ctx.JSON(response.Success(types.MePermissionsResponse{Roles: roles, Permissions: entries}))
}

// MeCommander godoc
// @Summary     Get commander profile
// @Tags        Me
// @Produce     json
// @Success     200  {object}  MeCommanderResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     403  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/me/commander [get]
func (handler *MeHandler) Commander(ctx iris.Context) {
	commander, _, ok := loadCommanderForUser(ctx)
	if !ok {
		return
	}
	payload := types.MeCommanderResponse{
		CommanderID: commander.CommanderID,
		Name:        commander.Name,
		Level:       uint32(commander.Level),
	}
	_ = ctx.JSON(response.Success(payload))
}

// MeResources godoc
// @Summary     Get user resources
// @Tags        Me
// @Produce     json
// @Success     200  {object}  PlayerResourcesResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     403  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/me/resources [get]
func (handler *MeHandler) Resources(ctx iris.Context) {
	commander, _, ok := loadCommanderForUser(ctx)
	if !ok {
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

// MeUpdateResources godoc
// @Summary     Update user resources
// @Tags        Me
// @Accept      json
// @Produce     json
// @Param       payload  body  types.ResourceUpdateRequest  true  "Resource update"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     403  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/me/resources [put]
func (handler *MeHandler) UpdateResources(ctx iris.Context) {
	commander, user, ok := loadCommanderForUser(ctx)
	if !ok {
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
	auth.LogUserAudit("self.resources.update", &user.ID, user.CommanderID, map[string]interface{}{"count": len(req.Resources)})
	_ = ctx.JSON(response.Success(nil))
}

// MeGiveShip godoc
// @Summary     Give ship to user
// @Tags        Me
// @Accept      json
// @Produce     json
// @Param       payload  body  types.GiveShipRequest  true  "Ship grant"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     403  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/me/give-ship [post]
func (handler *MeHandler) GiveShip(ctx iris.Context) {
	commander, user, ok := loadCommanderForUser(ctx)
	if !ok {
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
	auth.LogUserAudit("self.ships.give", &user.ID, user.CommanderID, map[string]interface{}{"ship_id": req.ShipID})
	_ = ctx.JSON(response.Success(nil))
}

// MeGiveItem godoc
// @Summary     Give item to user
// @Tags        Me
// @Accept      json
// @Produce     json
// @Param       payload  body  types.GiveItemRequest  true  "Item grant"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     403  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/me/give-item [post]
func (handler *MeHandler) GiveItem(ctx iris.Context) {
	commander, user, ok := loadCommanderForUser(ctx)
	if !ok {
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
	auth.LogUserAudit("self.items.give", &user.ID, user.CommanderID, map[string]interface{}{"item_id": req.ItemID, "amount": req.Amount})
	_ = ctx.JSON(response.Success(nil))
}

// MeGiveSkin godoc
// @Summary     Give skin to user
// @Tags        Me
// @Accept      json
// @Produce     json
// @Param       payload  body  types.GiveSkinRequest  true  "Skin grant"
// @Success     204  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     401  {object}  APIErrorResponseDoc
// @Failure     403  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/me/give-skin [post]
func (handler *MeHandler) GiveSkin(ctx iris.Context) {
	commander, user, ok := loadCommanderForUser(ctx)
	if !ok {
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
	auth.LogUserAudit("self.skins.give", &user.ID, user.CommanderID, map[string]interface{}{"skin_id": req.SkinID})
	ctx.StatusCode(iris.StatusNoContent)
}

func loadCommanderForUser(ctx iris.Context) (orm.Commander, *orm.Account, bool) {
	user, ok := middleware.GetAccount(ctx)
	if !ok {
		ctx.StatusCode(iris.StatusUnauthorized)
		_ = ctx.JSON(response.Error("auth.session_missing", "session required", nil))
		return orm.Commander{}, nil, false
	}
	if user.CommanderID == nil {
		ctx.StatusCode(iris.StatusForbidden)
		_ = ctx.JSON(response.Error("permissions.denied", "permission denied", nil))
		return orm.Commander{}, nil, false
	}
	commander, err := loadCommanderDetailByID(*user.CommanderID)
	if err != nil {
		writeCommanderError(ctx, err)
		return orm.Commander{}, nil, false
	}
	return commander, user, true
}

func loadCommanderDetailByID(commanderID uint32) (orm.Commander, error) {
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
