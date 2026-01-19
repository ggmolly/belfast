package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type GameDataHandler struct{}

func NewGameDataHandler() *GameDataHandler {
	return &GameDataHandler{}
}

func RegisterGameDataRoutes(party iris.Party, handler *GameDataHandler) {
	party.Get("/ships", handler.ListShips)
	party.Get("/ships/{id:uint}", handler.ShipDetail)
	party.Get("/ships/{id:uint}/skins", handler.ShipSkins)
	party.Get("/items", handler.ListItems)
	party.Get("/items/{id:uint}", handler.ItemDetail)
	party.Get("/resources", handler.ListResources)
	party.Get("/resources/{id:uint}", handler.ResourceDetail)
	party.Get("/skins", handler.ListSkins)
	party.Get("/skins/{id:uint}", handler.SkinDetail)
}

func (handler *GameDataHandler) ListShips(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	params := orm.ShipQueryParams{
		Offset: pagination.Offset,
		Limit:  pagination.Limit,
		Name:   strings.TrimSpace(ctx.URLParam("name")),
	}

	rarityID, err := parseOptionalUint32(ctx.URLParam("rarity"), "rarity")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	typeID, err := parseOptionalUint32(ctx.URLParam("type"), "type")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	nationalityID, err := parseOptionalUint32(ctx.URLParam("nationality"), "nationality")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	params.RarityID = rarityID
	params.TypeID = typeID
	params.NationalityID = nationalityID

	result, err := orm.ListShips(orm.GormDB, params)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list ships", nil))
		return
	}

	ships := make([]types.ShipSummary, 0, len(result.Ships))
	for _, ship := range result.Ships {
		ships = append(ships, types.ShipSummary{
			ID:          ship.TemplateID,
			Name:        ship.Name,
			RarityID:    ship.RarityID,
			Star:        ship.Star,
			Type:        ship.Type,
			Nationality: ship.Nationality,
			BuildTime:   ship.BuildTime,
			PoolID:      ship.PoolID,
		})
	}

	payload := types.ShipListResponse{
		Ships: ships,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  result.Total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) ShipDetail(ctx iris.Context) {
	shipID, err := parsePathUint32(ctx.Params().Get("id"), "ship id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var ship orm.Ship
	if err := orm.GormDB.First(&ship, "template_id = ?", shipID).Error; err != nil {
		writeGameDataError(ctx, err, "ship")
		return
	}

	payload := types.ShipSummary{
		ID:          ship.TemplateID,
		Name:        ship.Name,
		RarityID:    ship.RarityID,
		Star:        ship.Star,
		Type:        ship.Type,
		Nationality: ship.Nationality,
		BuildTime:   ship.BuildTime,
		PoolID:      ship.PoolID,
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) ListItems(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result, err := orm.ListItems(orm.GormDB, orm.ItemQueryParams{Offset: pagination.Offset, Limit: pagination.Limit})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list items", nil))
		return
	}

	items := make([]types.ItemSummary, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, types.ItemSummary{
			ID:          item.ID,
			Name:        item.Name,
			Rarity:      item.Rarity,
			ShopID:      item.ShopID,
			Type:        item.Type,
			VirtualType: item.VirtualType,
		})
	}

	payload := types.ItemListResponse{
		Items: items,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  result.Total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) ItemDetail(ctx iris.Context) {
	itemID, err := parsePathUint32(ctx.Params().Get("id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var item orm.Item
	if err := orm.GormDB.First(&item, itemID).Error; err != nil {
		writeGameDataError(ctx, err, "item")
		return
	}

	payload := types.ItemSummary{
		ID:          item.ID,
		Name:        item.Name,
		Rarity:      item.Rarity,
		ShopID:      item.ShopID,
		Type:        item.Type,
		VirtualType: item.VirtualType,
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) ListResources(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result, err := orm.ListResources(orm.GormDB, orm.ResourceQueryParams{Offset: pagination.Offset, Limit: pagination.Limit})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list resources", nil))
		return
	}

	resources := make([]types.ResourceSummary, 0, len(result.Resources))
	for _, res := range result.Resources {
		resources = append(resources, types.ResourceSummary{
			ID:     res.ID,
			ItemID: res.ItemID,
			Name:   res.Name,
		})
	}

	payload := types.ResourceListResponse{
		Resources: resources,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  result.Total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) ResourceDetail(ctx iris.Context) {
	resourceID, err := parsePathUint32(ctx.Params().Get("id"), "resource id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var resource orm.Resource
	if err := orm.GormDB.First(&resource, resourceID).Error; err != nil {
		writeGameDataError(ctx, err, "resource")
		return
	}

	payload := types.ResourceSummary{
		ID:     resource.ID,
		ItemID: resource.ItemID,
		Name:   resource.Name,
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) ListSkins(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result, err := orm.ListSkins(orm.GormDB, orm.SkinQueryParams{Offset: pagination.Offset, Limit: pagination.Limit})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list skins", nil))
		return
	}

	skins := make([]types.SkinSummary, 0, len(result.Skins))
	for _, skin := range result.Skins {
		skins = append(skins, types.SkinSummary{
			ID:        skin.ID,
			Name:      skin.Name,
			ShipGroup: skin.ShipGroup,
		})
	}

	payload := types.SkinListResponse{
		Skins: skins,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  result.Total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) SkinDetail(ctx iris.Context) {
	skinID, err := parsePathUint32(ctx.Params().Get("id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var skin orm.Skin
	if err := orm.GormDB.First(&skin, skinID).Error; err != nil {
		writeGameDataError(ctx, err, "skin")
		return
	}

	payload := types.SkinSummary{
		ID:        skin.ID,
		Name:      skin.Name,
		ShipGroup: skin.ShipGroup,
	}

	_ = ctx.JSON(response.Success(payload))
}

func (handler *GameDataHandler) ShipSkins(ctx iris.Context) {
	shipID, err := parsePathUint32(ctx.Params().Get("id"), "ship id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result, err := orm.ListSkinsByShipGroup(orm.GormDB, shipID, orm.SkinQueryParams{Offset: pagination.Offset, Limit: pagination.Limit})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list ship skins", nil))
		return
	}

	skins := make([]types.SkinSummary, 0, len(result.Skins))
	for _, skin := range result.Skins {
		skins = append(skins, types.SkinSummary{
			ID:        skin.ID,
			Name:      skin.Name,
			ShipGroup: skin.ShipGroup,
		})
	}

	payload := types.SkinListResponse{
		Skins: skins,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  result.Total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

func parseOptionalUint32(value string, name string) (*uint32, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return nil, fmt.Errorf("invalid %s", name)
	}
	result := uint32(parsed)
	return &result, nil
}

func parsePathUint32(value string, name string) (uint32, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return uint32(parsed), nil
}

func writeGameDataError(ctx iris.Context, err error, item string) {
	if err == gorm.ErrRecordNotFound {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", fmt.Sprintf("%s not found", item), nil))
		return
	}
	ctx.StatusCode(iris.StatusInternalServerError)
	_ = ctx.JSON(response.Error("internal_error", fmt.Sprintf("failed to load %s", item), nil))
}
