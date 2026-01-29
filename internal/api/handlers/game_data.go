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
	party.Post("/ships", handler.CreateShip)
	party.Get("/ships/{id:uint}", handler.ShipDetail)
	party.Put("/ships/{id:uint}", handler.UpdateShip)
	party.Delete("/ships/{id:uint}", handler.DeleteShip)
	party.Get("/ships/{id:uint}/skins", handler.ShipSkins)
	party.Get("/requisition-ships", handler.ListRequisitionShips)
	party.Post("/requisition-ships", handler.CreateRequisitionShip)
	party.Delete("/requisition-ships/{ship_id:uint}", handler.DeleteRequisitionShip)
	party.Get("/equipment", handler.ListEquipment)
	party.Post("/equipment", handler.CreateEquipment)
	party.Get("/equipment/{id:uint}", handler.EquipmentDetail)
	party.Put("/equipment/{id:uint}", handler.UpdateEquipment)
	party.Delete("/equipment/{id:uint}", handler.DeleteEquipment)
	party.Get("/weapons", handler.ListWeapons)
	party.Post("/weapons", handler.CreateWeapon)
	party.Get("/weapons/{id:uint}", handler.WeaponDetail)
	party.Put("/weapons/{id:uint}", handler.UpdateWeapon)
	party.Delete("/weapons/{id:uint}", handler.DeleteWeapon)
	party.Get("/skills", handler.ListSkills)
	party.Post("/skills", handler.CreateSkill)
	party.Get("/skills/{id:uint}", handler.SkillDetail)
	party.Put("/skills/{id:uint}", handler.UpdateSkill)
	party.Delete("/skills/{id:uint}", handler.DeleteSkill)
	party.Get("/buffs", handler.ListBuffs)
	party.Post("/buffs", handler.CreateBuff)
	party.Get("/buffs/{id:uint}", handler.BuffDetail)
	party.Put("/buffs/{id:uint}", handler.UpdateBuff)
	party.Delete("/buffs/{id:uint}", handler.DeleteBuff)
	party.Get("/ship-types", handler.ListShipTypes)
	party.Post("/ship-types", handler.CreateShipType)
	party.Get("/ship-types/{id:uint}", handler.ShipTypeDetail)
	party.Put("/ship-types/{id:uint}", handler.UpdateShipType)
	party.Delete("/ship-types/{id:uint}", handler.DeleteShipType)
	party.Get("/rarities", handler.ListRarities)
	party.Post("/rarities", handler.CreateRarity)
	party.Get("/rarities/{id:uint}", handler.RarityDetail)
	party.Put("/rarities/{id:uint}", handler.UpdateRarity)
	party.Delete("/rarities/{id:uint}", handler.DeleteRarity)
	party.Get("/items", handler.ListItems)
	party.Post("/items", handler.CreateItem)
	party.Get("/items/{id:uint}", handler.ItemDetail)
	party.Put("/items/{id:uint}", handler.UpdateItem)
	party.Delete("/items/{id:uint}", handler.DeleteItem)
	party.Get("/resources", handler.ListResources)
	party.Post("/resources", handler.CreateResource)
	party.Get("/resources/{id:uint}", handler.ResourceDetail)
	party.Put("/resources/{id:uint}", handler.UpdateResource)
	party.Delete("/resources/{id:uint}", handler.DeleteResource)
	party.Get("/skins", handler.ListSkins)
	party.Post("/skins", handler.CreateSkin)
	party.Get("/skins/{id:uint}", handler.SkinDetail)
	party.Put("/skins/{id:uint}", handler.UpdateSkin)
	party.Delete("/skins/{id:uint}", handler.DeleteSkin)
	party.Get("/skin-restrictions", handler.ListSkinRestrictions)
	party.Post("/skin-restrictions", handler.CreateSkinRestriction)
	party.Get("/skin-restrictions/{skin_id:uint}", handler.SkinRestrictionDetail)
	party.Put("/skin-restrictions/{skin_id:uint}", handler.UpdateSkinRestriction)
	party.Delete("/skin-restrictions/{skin_id:uint}", handler.DeleteSkinRestriction)
	party.Get("/skin-restriction-windows", handler.ListSkinRestrictionWindows)
	party.Post("/skin-restriction-windows", handler.CreateSkinRestrictionWindow)
	party.Get("/skin-restriction-windows/{id:uint}", handler.SkinRestrictionWindowDetail)
	party.Put("/skin-restriction-windows/{id:uint}", handler.UpdateSkinRestrictionWindow)
	party.Delete("/skin-restriction-windows/{id:uint}", handler.DeleteSkinRestrictionWindow)
	party.Get("/config-entries", handler.ListConfigEntries)
	party.Post("/config-entries", handler.CreateConfigEntry)
	party.Get("/config-entries/{id:uint}", handler.ConfigEntryDetail)
	party.Put("/config-entries/{id:uint}", handler.UpdateConfigEntry)
	party.Delete("/config-entries/{id:uint}", handler.DeleteConfigEntry)
	party.Get("/livingarea-covers", handler.ListLivingAreaCovers)
	party.Get("/attire/icon-frames", handler.ListIconFrames)
	party.Get("/attire/chat-frames", handler.ListChatFrames)
	party.Get("/attire/battle-ui", handler.ListBattleUIStyles)
}

// ListShips godoc
// @Summary     List ships
// @Description List ships with optional filters
// @Tags        Ships
// @Produce     json
// @Param       offset       query     int     false  "Pagination offset"
// @Param       limit        query     int     false  "Pagination limit"
// @Param       rarity       query     int     false  "Filter by rarity"
// @Param       type         query     int     false  "Filter by ship type"
// @Param       nationality  query     int     false  "Filter by nationality"
// @Param       name         query     string  false  "Filter by name"
// @Success     200  {object}  ListShipsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ships [get]
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
			EnglishName: ship.EnglishName,
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

// ShipDetail godoc
// @Summary     Get ship details
// @Tags        Ships
// @Produce     json
// @Param       id   path      int  true  "Ship ID"
// @Success     200  {object}  ShipSummaryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ships/{id} [get]
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
		EnglishName: ship.EnglishName,
		RarityID:    ship.RarityID,
		Star:        ship.Star,
		Type:        ship.Type,
		Nationality: ship.Nationality,
		BuildTime:   ship.BuildTime,
		PoolID:      ship.PoolID,
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateShip godoc
// @Summary     Create ship
// @Tags        Ships
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ShipSummary  true  "Ship"
// @Success     200  {object}  ShipMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ships [post]
func (handler *GameDataHandler) CreateShip(ctx iris.Context) {
	var req types.ShipSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	englishName := strings.TrimSpace(req.EnglishName)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}
	if englishName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "english_name is required", nil))
		return
	}

	ship := orm.Ship{
		TemplateID:  req.ID,
		Name:        name,
		EnglishName: englishName,
		RarityID:    req.RarityID,
		Star:        req.Star,
		Type:        req.Type,
		Nationality: req.Nationality,
		BuildTime:   req.BuildTime,
		PoolID:      req.PoolID,
	}

	if err := orm.GormDB.Create(&ship).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create ship", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateShip godoc
// @Summary     Update ship
// @Tags        Ships
// @Accept      json
// @Produce     json
// @Param       id       path      int               true  "Ship ID"
// @Param       payload  body      types.ShipSummary  true  "Ship"
// @Success     200  {object}  ShipMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ships/{id} [put]
func (handler *GameDataHandler) UpdateShip(ctx iris.Context) {
	shipID, err := parsePathUint32(ctx.Params().Get("id"), "ship id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.ShipSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	englishName := strings.TrimSpace(req.EnglishName)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}
	if englishName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "english_name is required", nil))
		return
	}

	var ship orm.Ship
	if err := orm.GormDB.First(&ship, "template_id = ?", shipID).Error; err != nil {
		writeGameDataError(ctx, err, "ship")
		return
	}

	ship.Name = name
	ship.EnglishName = englishName
	ship.RarityID = req.RarityID
	ship.Star = req.Star
	ship.Type = req.Type
	ship.Nationality = req.Nationality
	ship.BuildTime = req.BuildTime
	ship.PoolID = req.PoolID

	if err := orm.GormDB.Save(&ship).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update ship", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteShip godoc
// @Summary     Delete ship
// @Tags        Ships
// @Produce     json
// @Param       id   path      int  true  "Ship ID"
// @Success     200  {object}  ShipMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ships/{id} [delete]
func (handler *GameDataHandler) DeleteShip(ctx iris.Context) {
	shipID, err := parsePathUint32(ctx.Params().Get("id"), "ship id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Ship{}, "template_id = ?", shipID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete ship", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "ship not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListRequisitionShips godoc
// @Summary     List requisition ships
// @Tags        RequisitionShips
// @Produce     json
// @Success     200  {object}  RequisitionShipListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/requisition-ships [get]
func (handler *GameDataHandler) ListRequisitionShips(ctx iris.Context) {
	ids, err := orm.ListRequisitionShipIDs()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list requisition ships", nil))
		return
	}

	payload := types.RequisitionShipListResponse{ShipIDs: ids}
	_ = ctx.JSON(response.Success(payload))
}

// CreateRequisitionShip godoc
// @Summary     Create requisition ship
// @Tags        RequisitionShips
// @Accept      json
// @Produce     json
// @Param       payload  body      types.RequisitionShipRequest  true  "Requisition ship"
// @Success     200  {object}  RequisitionShipMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/requisition-ships [post]
func (handler *GameDataHandler) CreateRequisitionShip(ctx iris.Context) {
	var req types.RequisitionShipRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	if req.ShipID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "ship_id is required", nil))
		return
	}

	entry := orm.RequisitionShip{ShipID: req.ShipID}
	if err := orm.GormDB.Create(&entry).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create requisition ship", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteRequisitionShip godoc
// @Summary     Delete requisition ship
// @Tags        RequisitionShips
// @Produce     json
// @Param       ship_id  path      int  true  "Ship ID"
// @Success     200  {object}  RequisitionShipMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/requisition-ships/{ship_id} [delete]
func (handler *GameDataHandler) DeleteRequisitionShip(ctx iris.Context) {
	shipID, err := parsePathUint32(ctx.Params().Get("ship_id"), "ship id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.RequisitionShip{}, "ship_id = ?", shipID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete requisition ship", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "requisition ship not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListShipTypes godoc
// @Summary     List ship types
// @Tags        ShipTypes
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListShipTypesResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ship-types [get]
func (handler *GameDataHandler) ListShipTypes(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.ShipType{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list ship types", nil))
		return
	}

	var shipTypes []orm.ShipType
	query = query.Order("id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&shipTypes).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list ship types", nil))
		return
	}

	results := make([]types.ShipTypeSummary, 0, len(shipTypes))
	for _, shipType := range shipTypes {
		results = append(results, types.ShipTypeSummary{
			ID:   shipType.ID,
			Name: shipType.Name,
		})
	}

	payload := types.ShipTypeListResponse{
		ShipTypes: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// ShipTypeDetail godoc
// @Summary     Get ship type details
// @Tags        ShipTypes
// @Produce     json
// @Param       id   path      int  true  "Ship type ID"
// @Success     200  {object}  ShipTypeSummaryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ship-types/{id} [get]
func (handler *GameDataHandler) ShipTypeDetail(ctx iris.Context) {
	shipTypeID, err := parsePathUint32(ctx.Params().Get("id"), "ship type id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var shipType orm.ShipType
	if err := orm.GormDB.First(&shipType, shipTypeID).Error; err != nil {
		writeGameDataError(ctx, err, "ship type")
		return
	}

	payload := types.ShipTypeSummary{
		ID:   shipType.ID,
		Name: shipType.Name,
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateShipType godoc
// @Summary     Create ship type
// @Tags        ShipTypes
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ShipTypeSummary  true  "Ship type"
// @Success     200  {object}  ShipTypeMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ship-types [post]
func (handler *GameDataHandler) CreateShipType(ctx iris.Context) {
	var req types.ShipTypeSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	shipType := orm.ShipType{
		ID:   req.ID,
		Name: name,
	}

	if err := orm.GormDB.Create(&shipType).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create ship type", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateShipType godoc
// @Summary     Update ship type
// @Tags        ShipTypes
// @Accept      json
// @Produce     json
// @Param       id       path      int                    true  "Ship type ID"
// @Param       payload  body      types.ShipTypeSummary  true  "Ship type"
// @Success     200  {object}  ShipTypeMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ship-types/{id} [put]
func (handler *GameDataHandler) UpdateShipType(ctx iris.Context) {
	shipTypeID, err := parsePathUint32(ctx.Params().Get("id"), "ship type id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.ShipTypeSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	var shipType orm.ShipType
	if err := orm.GormDB.First(&shipType, shipTypeID).Error; err != nil {
		writeGameDataError(ctx, err, "ship type")
		return
	}

	shipType.Name = name

	if err := orm.GormDB.Save(&shipType).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update ship type", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteShipType godoc
// @Summary     Delete ship type
// @Tags        ShipTypes
// @Produce     json
// @Param       id   path      int  true  "Ship type ID"
// @Success     200  {object}  ShipTypeMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ship-types/{id} [delete]
func (handler *GameDataHandler) DeleteShipType(ctx iris.Context) {
	shipTypeID, err := parsePathUint32(ctx.Params().Get("id"), "ship type id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.ShipType{}, shipTypeID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete ship type", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "ship type not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListRarities godoc
// @Summary     List rarities
// @Tags        Rarities
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListRaritiesResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/rarities [get]
func (handler *GameDataHandler) ListRarities(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.Rarity{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list rarities", nil))
		return
	}

	var rarities []orm.Rarity
	query = query.Order("id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&rarities).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list rarities", nil))
		return
	}

	results := make([]types.RaritySummary, 0, len(rarities))
	for _, rarity := range rarities {
		results = append(results, types.RaritySummary{
			ID:   rarity.ID,
			Name: rarity.Name,
		})
	}

	payload := types.RarityListResponse{
		Rarities: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// RarityDetail godoc
// @Summary     Get rarity details
// @Tags        Rarities
// @Produce     json
// @Param       id   path      int  true  "Rarity ID"
// @Success     200  {object}  RaritySummaryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/rarities/{id} [get]
func (handler *GameDataHandler) RarityDetail(ctx iris.Context) {
	rarityID, err := parsePathUint32(ctx.Params().Get("id"), "rarity id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var rarity orm.Rarity
	if err := orm.GormDB.First(&rarity, rarityID).Error; err != nil {
		writeGameDataError(ctx, err, "rarity")
		return
	}

	payload := types.RaritySummary{
		ID:   rarity.ID,
		Name: rarity.Name,
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateRarity godoc
// @Summary     Create rarity
// @Tags        Rarities
// @Accept      json
// @Produce     json
// @Param       payload  body      types.RaritySummary  true  "Rarity"
// @Success     200  {object}  RarityMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/rarities [post]
func (handler *GameDataHandler) CreateRarity(ctx iris.Context) {
	var req types.RaritySummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	rarity := orm.Rarity{
		ID:   req.ID,
		Name: name,
	}

	if err := orm.GormDB.Create(&rarity).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create rarity", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateRarity godoc
// @Summary     Update rarity
// @Tags        Rarities
// @Accept      json
// @Produce     json
// @Param       id       path      int                  true  "Rarity ID"
// @Param       payload  body      types.RaritySummary  true  "Rarity"
// @Success     200  {object}  RarityMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/rarities/{id} [put]
func (handler *GameDataHandler) UpdateRarity(ctx iris.Context) {
	rarityID, err := parsePathUint32(ctx.Params().Get("id"), "rarity id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.RaritySummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	var rarity orm.Rarity
	if err := orm.GormDB.First(&rarity, rarityID).Error; err != nil {
		writeGameDataError(ctx, err, "rarity")
		return
	}

	rarity.Name = name

	if err := orm.GormDB.Save(&rarity).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update rarity", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteRarity godoc
// @Summary     Delete rarity
// @Tags        Rarities
// @Produce     json
// @Param       id   path      int  true  "Rarity ID"
// @Success     200  {object}  RarityMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/rarities/{id} [delete]
func (handler *GameDataHandler) DeleteRarity(ctx iris.Context) {
	rarityID, err := parsePathUint32(ctx.Params().Get("id"), "rarity id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Rarity{}, rarityID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete rarity", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "rarity not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListItems godoc
// @Summary     List items
// @Tags        Items
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListItemsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/items [get]
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

// ItemDetail godoc
// @Summary     Get item details
// @Tags        Items
// @Produce     json
// @Param       id   path      int  true  "Item ID"
// @Success     200  {object}  ItemSummaryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/items/{id} [get]
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

// CreateItem godoc
// @Summary     Create item
// @Tags        Items
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ItemCreateRequest  true  "Item"
// @Success     200  {object}  ItemMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/items [post]
func (handler *GameDataHandler) CreateItem(ctx iris.Context) {
	var req types.ItemCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	item := orm.Item{
		ID:          req.ID,
		Name:        name,
		Rarity:      req.Rarity,
		ShopID:      req.ShopID,
		Type:        req.Type,
		VirtualType: req.VirtualType,
	}

	if err := orm.GormDB.Create(&item).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create item", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateItem godoc
// @Summary     Update item
// @Tags        Items
// @Accept      json
// @Produce     json
// @Param       id       path      int                     true  "Item ID"
// @Param       payload  body      types.ItemUpdateRequest  true  "Item"
// @Success     200  {object}  ItemMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/items/{id} [put]
func (handler *GameDataHandler) UpdateItem(ctx iris.Context) {
	itemID, err := parsePathUint32(ctx.Params().Get("id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.ItemUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	var item orm.Item
	if err := orm.GormDB.First(&item, itemID).Error; err != nil {
		writeGameDataError(ctx, err, "item")
		return
	}

	item.Name = name
	item.Rarity = req.Rarity
	item.ShopID = req.ShopID
	item.Type = req.Type
	item.VirtualType = req.VirtualType

	if err := orm.GormDB.Save(&item).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update item", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteItem godoc
// @Summary     Delete item
// @Tags        Items
// @Produce     json
// @Param       id   path      int  true  "Item ID"
// @Success     200  {object}  ItemMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/items/{id} [delete]
func (handler *GameDataHandler) DeleteItem(ctx iris.Context) {
	itemID, err := parsePathUint32(ctx.Params().Get("id"), "item id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Item{}, itemID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete item", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "item not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListResources godoc
// @Summary     List resources
// @Tags        Resources
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListResourcesResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/resources [get]
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

// ResourceDetail godoc
// @Summary     Get resource details
// @Tags        Resources
// @Produce     json
// @Param       id   path      int  true  "Resource ID"
// @Success     200  {object}  ResourceSummaryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/resources/{id} [get]
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

// CreateResource godoc
// @Summary     Create resource
// @Tags        Resources
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ResourceCreateRequest  true  "Resource"
// @Success     200  {object}  ResourceMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/resources [post]
func (handler *GameDataHandler) CreateResource(ctx iris.Context) {
	var req types.ResourceCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	resource := orm.Resource{
		ID:     req.ID,
		ItemID: req.ItemID,
		Name:   name,
	}

	if err := orm.GormDB.Create(&resource).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create resource", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateResource godoc
// @Summary     Update resource
// @Tags        Resources
// @Accept      json
// @Produce     json
// @Param       id       path      int                        true  "Resource ID"
// @Param       payload  body      types.ResourceUpdatePayload  true  "Resource"
// @Success     200  {object}  ResourceMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/resources/{id} [put]
func (handler *GameDataHandler) UpdateResource(ctx iris.Context) {
	resourceID, err := parsePathUint32(ctx.Params().Get("id"), "resource id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.ResourceUpdatePayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	var resource orm.Resource
	if err := orm.GormDB.First(&resource, resourceID).Error; err != nil {
		writeGameDataError(ctx, err, "resource")
		return
	}

	resource.Name = name
	resource.ItemID = req.ItemID

	if err := orm.GormDB.Save(&resource).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update resource", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteResource godoc
// @Summary     Delete resource
// @Tags        Resources
// @Produce     json
// @Param       id   path      int  true  "Resource ID"
// @Success     200  {object}  ResourceMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/resources/{id} [delete]
func (handler *GameDataHandler) DeleteResource(ctx iris.Context) {
	resourceID, err := parsePathUint32(ctx.Params().Get("id"), "resource id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Resource{}, resourceID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete resource", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "resource not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListEquipment godoc
// @Summary     List equipment
// @Tags        Equipment
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListEquipmentResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/equipment [get]
func (handler *GameDataHandler) ListEquipment(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.Equipment{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list equipment", nil))
		return
	}

	var equipment []orm.Equipment
	query = query.Order("id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&equipment).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list equipment", nil))
		return
	}

	results := make([]types.EquipmentPayload, 0, len(equipment))
	for _, entry := range equipment {
		results = append(results, equipmentPayloadFromModel(entry))
	}

	payload := types.EquipmentListResponse{
		Equipment: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// EquipmentDetail godoc
// @Summary     Get equipment details
// @Tags        Equipment
// @Produce     json
// @Param       id   path      int  true  "Equipment ID"
// @Success     200  {object}  EquipmentDetailResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/equipment/{id} [get]
func (handler *GameDataHandler) EquipmentDetail(ctx iris.Context) {
	equipmentID, err := parsePathUint32(ctx.Params().Get("id"), "equipment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var equipment orm.Equipment
	if err := orm.GormDB.First(&equipment, equipmentID).Error; err != nil {
		writeGameDataError(ctx, err, "equipment")
		return
	}

	payload := equipmentPayloadFromModel(equipment)

	_ = ctx.JSON(response.Success(payload))
}

// CreateEquipment godoc
// @Summary     Create equipment
// @Tags        Equipment
// @Accept      json
// @Produce     json
// @Param       payload  body      types.EquipmentPayload  true  "Equipment"
// @Success     200  {object}  EquipmentMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/equipment [post]
func (handler *GameDataHandler) CreateEquipment(ctx iris.Context) {
	var req types.EquipmentPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}

	equipment := orm.Equipment{ID: req.ID}
	applyEquipmentPayload(&equipment, req)

	if err := orm.GormDB.Create(&equipment).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create equipment", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateEquipment godoc
// @Summary     Update equipment
// @Tags        Equipment
// @Accept      json
// @Produce     json
// @Param       id       path      int                    true  "Equipment ID"
// @Param       payload  body      types.EquipmentPayload  true  "Equipment"
// @Success     200  {object}  EquipmentMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/equipment/{id} [put]
func (handler *GameDataHandler) UpdateEquipment(ctx iris.Context) {
	equipmentID, err := parsePathUint32(ctx.Params().Get("id"), "equipment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.EquipmentPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	var equipment orm.Equipment
	if err := orm.GormDB.First(&equipment, equipmentID).Error; err != nil {
		writeGameDataError(ctx, err, "equipment")
		return
	}

	applyEquipmentPayload(&equipment, req)

	if err := orm.GormDB.Save(&equipment).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update equipment", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteEquipment godoc
// @Summary     Delete equipment
// @Tags        Equipment
// @Produce     json
// @Param       id   path      int  true  "Equipment ID"
// @Success     200  {object}  EquipmentMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/equipment/{id} [delete]
func (handler *GameDataHandler) DeleteEquipment(ctx iris.Context) {
	equipmentID, err := parsePathUint32(ctx.Params().Get("id"), "equipment id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Equipment{}, equipmentID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete equipment", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "equipment not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListWeapons godoc
// @Summary     List weapons
// @Tags        Weapons
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListWeaponsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/weapons [get]
func (handler *GameDataHandler) ListWeapons(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.Weapon{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list weapons", nil))
		return
	}

	var weapons []orm.Weapon
	query = query.Order("id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&weapons).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list weapons", nil))
		return
	}

	results := make([]types.WeaponPayload, 0, len(weapons))
	for _, entry := range weapons {
		results = append(results, weaponPayloadFromModel(entry))
	}

	payload := types.WeaponListResponse{
		Weapons: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// WeaponDetail godoc
// @Summary     Get weapon details
// @Tags        Weapons
// @Produce     json
// @Param       id   path      int  true  "Weapon ID"
// @Success     200  {object}  WeaponDetailResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/weapons/{id} [get]
func (handler *GameDataHandler) WeaponDetail(ctx iris.Context) {
	weaponID, err := parsePathUint32(ctx.Params().Get("id"), "weapon id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var weapon orm.Weapon
	if err := orm.GormDB.First(&weapon, weaponID).Error; err != nil {
		writeGameDataError(ctx, err, "weapon")
		return
	}

	payload := weaponPayloadFromModel(weapon)

	_ = ctx.JSON(response.Success(payload))
}

// CreateWeapon godoc
// @Summary     Create weapon
// @Tags        Weapons
// @Accept      json
// @Produce     json
// @Param       payload  body      types.WeaponPayload  true  "Weapon"
// @Success     200  {object}  WeaponMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/weapons [post]
func (handler *GameDataHandler) CreateWeapon(ctx iris.Context) {
	var req types.WeaponPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}

	weapon := orm.Weapon{ID: req.ID}
	applyWeaponPayload(&weapon, req)

	if err := orm.GormDB.Create(&weapon).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create weapon", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateWeapon godoc
// @Summary     Update weapon
// @Tags        Weapons
// @Accept      json
// @Produce     json
// @Param       id       path      int                  true  "Weapon ID"
// @Param       payload  body      types.WeaponPayload  true  "Weapon"
// @Success     200  {object}  WeaponMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/weapons/{id} [put]
func (handler *GameDataHandler) UpdateWeapon(ctx iris.Context) {
	weaponID, err := parsePathUint32(ctx.Params().Get("id"), "weapon id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.WeaponPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	var weapon orm.Weapon
	if err := orm.GormDB.First(&weapon, weaponID).Error; err != nil {
		writeGameDataError(ctx, err, "weapon")
		return
	}

	applyWeaponPayload(&weapon, req)

	if err := orm.GormDB.Save(&weapon).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update weapon", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteWeapon godoc
// @Summary     Delete weapon
// @Tags        Weapons
// @Produce     json
// @Param       id   path      int  true  "Weapon ID"
// @Success     200  {object}  WeaponMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/weapons/{id} [delete]
func (handler *GameDataHandler) DeleteWeapon(ctx iris.Context) {
	weaponID, err := parsePathUint32(ctx.Params().Get("id"), "weapon id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Weapon{}, weaponID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete weapon", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "weapon not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListSkills godoc
// @Summary     List skills
// @Tags        Skills
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListSkillsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skills [get]
func (handler *GameDataHandler) ListSkills(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.Skill{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list skills", nil))
		return
	}

	var skills []orm.Skill
	query = query.Order("id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&skills).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list skills", nil))
		return
	}

	results := make([]types.SkillPayload, 0, len(skills))
	for _, skill := range skills {
		results = append(results, skillPayloadFromModel(skill))
	}

	payload := types.SkillListResponse{
		Skills: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// SkillDetail godoc
// @Summary     Get skill details
// @Tags        Skills
// @Produce     json
// @Param       id   path      int  true  "Skill ID"
// @Success     200  {object}  SkillDetailResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skills/{id} [get]
func (handler *GameDataHandler) SkillDetail(ctx iris.Context) {
	skillID, err := parsePathUint32(ctx.Params().Get("id"), "skill id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var skill orm.Skill
	if err := orm.GormDB.First(&skill, skillID).Error; err != nil {
		writeGameDataError(ctx, err, "skill")
		return
	}

	payload := skillPayloadFromModel(skill)

	_ = ctx.JSON(response.Success(payload))
}

// CreateSkill godoc
// @Summary     Create skill
// @Tags        Skills
// @Accept      json
// @Produce     json
// @Param       payload  body      types.SkillPayload  true  "Skill"
// @Success     200  {object}  SkillMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skills [post]
func (handler *GameDataHandler) CreateSkill(ctx iris.Context) {
	var req types.SkillPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	skill := orm.Skill{ID: req.ID}
	applySkillPayload(&skill, req)
	skill.Name = name

	if err := orm.GormDB.Create(&skill).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create skill", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateSkill godoc
// @Summary     Update skill
// @Tags        Skills
// @Accept      json
// @Produce     json
// @Param       id       path      int                 true  "Skill ID"
// @Param       payload  body      types.SkillPayload  true  "Skill"
// @Success     200  {object}  SkillMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skills/{id} [put]
func (handler *GameDataHandler) UpdateSkill(ctx iris.Context) {
	skillID, err := parsePathUint32(ctx.Params().Get("id"), "skill id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.SkillPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	var skill orm.Skill
	if err := orm.GormDB.First(&skill, skillID).Error; err != nil {
		writeGameDataError(ctx, err, "skill")
		return
	}

	applySkillPayload(&skill, req)
	skill.Name = name

	if err := orm.GormDB.Save(&skill).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update skill", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteSkill godoc
// @Summary     Delete skill
// @Tags        Skills
// @Produce     json
// @Param       id   path      int  true  "Skill ID"
// @Success     200  {object}  SkillMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skills/{id} [delete]
func (handler *GameDataHandler) DeleteSkill(ctx iris.Context) {
	skillID, err := parsePathUint32(ctx.Params().Get("id"), "skill id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Skill{}, skillID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete skill", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "skill not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListBuffs godoc
// @Summary     List buffs
// @Tags        Buffs
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListBuffsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/buffs [get]
func (handler *GameDataHandler) ListBuffs(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.Buff{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list buffs", nil))
		return
	}

	var buffs []orm.Buff
	query = query.Order("id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&buffs).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list buffs", nil))
		return
	}

	results := make([]types.BuffPayload, 0, len(buffs))
	for _, buff := range buffs {
		results = append(results, buffPayloadFromModel(buff))
	}

	payload := types.BuffListResponse{
		Buffs: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// BuffDetail godoc
// @Summary     Get buff details
// @Tags        Buffs
// @Produce     json
// @Param       id   path      int  true  "Buff ID"
// @Success     200  {object}  BuffDetailResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/buffs/{id} [get]
func (handler *GameDataHandler) BuffDetail(ctx iris.Context) {
	buffID, err := parsePathUint32(ctx.Params().Get("id"), "buff id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var buff orm.Buff
	if err := orm.GormDB.First(&buff, buffID).Error; err != nil {
		writeGameDataError(ctx, err, "buff")
		return
	}

	payload := buffPayloadFromModel(buff)

	_ = ctx.JSON(response.Success(payload))
}

// CreateBuff godoc
// @Summary     Create buff
// @Tags        Buffs
// @Accept      json
// @Produce     json
// @Param       payload  body      types.BuffPayload  true  "Buff"
// @Success     200  {object}  BuffMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/buffs [post]
func (handler *GameDataHandler) CreateBuff(ctx iris.Context) {
	var req types.BuffPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	benefitType := strings.TrimSpace(req.BenefitType)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}
	if benefitType == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "benefit_type is required", nil))
		return
	}

	buff := orm.Buff{
		ID:          req.ID,
		Name:        name,
		Description: strings.TrimSpace(req.Desc),
		MaxTime:     req.MaxTime,
		BenefitType: benefitType,
	}

	if err := orm.GormDB.Create(&buff).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create buff", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateBuff godoc
// @Summary     Update buff
// @Tags        Buffs
// @Accept      json
// @Produce     json
// @Param       id       path      int               true  "Buff ID"
// @Param       payload  body      types.BuffPayload  true  "Buff"
// @Success     200  {object}  BuffMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/buffs/{id} [put]
func (handler *GameDataHandler) UpdateBuff(ctx iris.Context) {
	buffID, err := parsePathUint32(ctx.Params().Get("id"), "buff id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.BuffPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	name := strings.TrimSpace(req.Name)
	benefitType := strings.TrimSpace(req.BenefitType)
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}
	if benefitType == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "benefit_type is required", nil))
		return
	}

	var buff orm.Buff
	if err := orm.GormDB.First(&buff, buffID).Error; err != nil {
		writeGameDataError(ctx, err, "buff")
		return
	}

	buff.Name = name
	buff.Description = strings.TrimSpace(req.Desc)
	buff.MaxTime = req.MaxTime
	buff.BenefitType = benefitType

	if err := orm.GormDB.Save(&buff).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update buff", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteBuff godoc
// @Summary     Delete buff
// @Tags        Buffs
// @Produce     json
// @Param       id   path      int  true  "Buff ID"
// @Success     200  {object}  BuffMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/buffs/{id} [delete]
func (handler *GameDataHandler) DeleteBuff(ctx iris.Context) {
	buffID, err := parsePathUint32(ctx.Params().Get("id"), "buff id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Buff{}, buffID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete buff", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "buff not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListSkins godoc
// @Summary     List skins
// @Tags        Skins
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListSkinsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skins [get]
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

// CreateSkin godoc
// @Summary     Create skin
// @Tags        Skins
// @Accept      json
// @Produce     json
// @Param       payload  body      types.SkinPayload  true  "Skin"
// @Success     200  {object}  SkinMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skins [post]
func (handler *GameDataHandler) CreateSkin(ctx iris.Context) {
	var req types.SkinPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if req.Name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	skin := orm.Skin{ID: req.ID}
	applySkinPayload(&skin, req)

	if err := orm.GormDB.Create(&skin).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create skin", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// SkinDetail godoc
// @Summary     Get skin details
// @Tags        Skins
// @Produce     json
// @Param       id   path      int  true  "Skin ID"
// @Success     200  {object}  SkinDetailResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skins/{id} [get]
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

	payload := skinPayloadFromModel(skin)

	_ = ctx.JSON(response.Success(payload))
}

// UpdateSkin godoc
// @Summary     Update skin
// @Tags        Skins
// @Accept      json
// @Produce     json
// @Param       id       path      int               true  "Skin ID"
// @Param       payload  body      types.SkinPayload  true  "Skin"
// @Success     200  {object}  SkinMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skins/{id} [put]
func (handler *GameDataHandler) UpdateSkin(ctx iris.Context) {
	skinID, err := parsePathUint32(ctx.Params().Get("id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.SkinPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}

	var skin orm.Skin
	if err := orm.GormDB.First(&skin, skinID).Error; err != nil {
		writeGameDataError(ctx, err, "skin")
		return
	}

	applySkinPayload(&skin, req)

	if err := orm.GormDB.Save(&skin).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update skin", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteSkin godoc
// @Summary     Delete skin
// @Tags        Skins
// @Produce     json
// @Param       id   path      int  true  "Skin ID"
// @Success     200  {object}  SkinMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skins/{id} [delete]
func (handler *GameDataHandler) DeleteSkin(ctx iris.Context) {
	skinID, err := parsePathUint32(ctx.Params().Get("id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Skin{}, skinID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete skin", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "skin not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListSkinRestrictions godoc
// @Summary     List skin restrictions
// @Tags        SkinRestrictions
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListSkinRestrictionsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restrictions [get]
func (handler *GameDataHandler) ListSkinRestrictions(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.GlobalSkinRestriction{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list skin restrictions", nil))
		return
	}

	var restrictions []orm.GlobalSkinRestriction
	query = query.Order("skin_id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&restrictions).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list skin restrictions", nil))
		return
	}

	results := make([]types.SkinRestrictionPayload, 0, len(restrictions))
	for _, restriction := range restrictions {
		results = append(results, types.SkinRestrictionPayload{
			SkinID: restriction.SkinID,
			Type:   restriction.Type,
		})
	}

	payload := types.SkinRestrictionListResponse{
		SkinRestrictions: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// SkinRestrictionDetail godoc
// @Summary     Get skin restriction details
// @Tags        SkinRestrictions
// @Produce     json
// @Param       skin_id   path      int  true  "Skin ID"
// @Success     200  {object}  SkinRestrictionResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restrictions/{skin_id} [get]
func (handler *GameDataHandler) SkinRestrictionDetail(ctx iris.Context) {
	skinID, err := parsePathUint32(ctx.Params().Get("skin_id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var restriction orm.GlobalSkinRestriction
	if err := orm.GormDB.First(&restriction, "skin_id = ?", skinID).Error; err != nil {
		writeGameDataError(ctx, err, "skin restriction")
		return
	}

	payload := types.SkinRestrictionPayload{
		SkinID: restriction.SkinID,
		Type:   restriction.Type,
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateSkinRestriction godoc
// @Summary     Create skin restriction
// @Tags        SkinRestrictions
// @Accept      json
// @Produce     json
// @Param       payload  body      types.SkinRestrictionCreateRequest  true  "Skin restriction"
// @Success     200  {object}  SkinRestrictionMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restrictions [post]
func (handler *GameDataHandler) CreateSkinRestriction(ctx iris.Context) {
	var req types.SkinRestrictionCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	if req.SkinID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "skin_id is required", nil))
		return
	}

	restriction := orm.GlobalSkinRestriction{
		SkinID: req.SkinID,
		Type:   req.Type,
	}

	if err := orm.GormDB.Create(&restriction).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create skin restriction", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateSkinRestriction godoc
// @Summary     Update skin restriction
// @Tags        SkinRestrictions
// @Accept      json
// @Produce     json
// @Param       skin_id   path      int  true  "Skin ID"
// @Param       payload  body      types.SkinRestrictionUpdateRequest  true  "Skin restriction"
// @Success     200  {object}  SkinRestrictionMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restrictions/{skin_id} [put]
func (handler *GameDataHandler) UpdateSkinRestriction(ctx iris.Context) {
	skinID, err := parsePathUint32(ctx.Params().Get("skin_id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.SkinRestrictionUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	var restriction orm.GlobalSkinRestriction
	if err := orm.GormDB.First(&restriction, "skin_id = ?", skinID).Error; err != nil {
		writeGameDataError(ctx, err, "skin restriction")
		return
	}

	restriction.Type = req.Type

	if err := orm.GormDB.Save(&restriction).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update skin restriction", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteSkinRestriction godoc
// @Summary     Delete skin restriction
// @Tags        SkinRestrictions
// @Produce     json
// @Param       skin_id   path      int  true  "Skin ID"
// @Success     200  {object}  SkinRestrictionMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restrictions/{skin_id} [delete]
func (handler *GameDataHandler) DeleteSkinRestriction(ctx iris.Context) {
	skinID, err := parsePathUint32(ctx.Params().Get("skin_id"), "skin id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.GlobalSkinRestriction{}, "skin_id = ?", skinID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete skin restriction", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "skin restriction not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListSkinRestrictionWindows godoc
// @Summary     List skin restriction windows
// @Tags        SkinRestrictionWindows
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ListSkinRestrictionWindowsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restriction-windows [get]
func (handler *GameDataHandler) ListSkinRestrictionWindows(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	query := orm.GormDB.Model(&orm.GlobalSkinRestrictionWindow{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list skin restriction windows", nil))
		return
	}

	var windows []orm.GlobalSkinRestrictionWindow
	query = query.Order("id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&windows).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list skin restriction windows", nil))
		return
	}

	results := make([]types.SkinRestrictionWindowPayload, 0, len(windows))
	for _, window := range windows {
		results = append(results, types.SkinRestrictionWindowPayload{
			ID:        window.ID,
			SkinID:    window.SkinID,
			Type:      window.Type,
			StartTime: window.StartTime,
			StopTime:  window.StopTime,
		})
	}

	payload := types.SkinRestrictionWindowListResponse{
		Windows: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// SkinRestrictionWindowDetail godoc
// @Summary     Get skin restriction window details
// @Tags        SkinRestrictionWindows
// @Produce     json
// @Param       id   path      int  true  "Window ID"
// @Success     200  {object}  SkinRestrictionWindowResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restriction-windows/{id} [get]
func (handler *GameDataHandler) SkinRestrictionWindowDetail(ctx iris.Context) {
	windowID, err := parsePathUint32(ctx.Params().Get("id"), "window id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var window orm.GlobalSkinRestrictionWindow
	if err := orm.GormDB.First(&window, windowID).Error; err != nil {
		writeGameDataError(ctx, err, "skin restriction window")
		return
	}

	payload := types.SkinRestrictionWindowPayload{
		ID:        window.ID,
		SkinID:    window.SkinID,
		Type:      window.Type,
		StartTime: window.StartTime,
		StopTime:  window.StopTime,
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateSkinRestrictionWindow godoc
// @Summary     Create skin restriction window
// @Tags        SkinRestrictionWindows
// @Accept      json
// @Produce     json
// @Param       payload  body      types.SkinRestrictionWindowCreateRequest  true  "Skin restriction window"
// @Success     200  {object}  SkinRestrictionWindowMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restriction-windows [post]
func (handler *GameDataHandler) CreateSkinRestrictionWindow(ctx iris.Context) {
	var req types.SkinRestrictionWindowCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if req.SkinID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "skin_id is required", nil))
		return
	}

	window := orm.GlobalSkinRestrictionWindow{
		ID:        req.ID,
		SkinID:    req.SkinID,
		Type:      req.Type,
		StartTime: req.StartTime,
		StopTime:  req.StopTime,
	}

	if err := orm.GormDB.Create(&window).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create skin restriction window", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateSkinRestrictionWindow godoc
// @Summary     Update skin restriction window
// @Tags        SkinRestrictionWindows
// @Accept      json
// @Produce     json
// @Param       id       path      int  true  "Window ID"
// @Param       payload  body      types.SkinRestrictionWindowUpdateRequest  true  "Skin restriction window"
// @Success     200  {object}  SkinRestrictionWindowMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restriction-windows/{id} [put]
func (handler *GameDataHandler) UpdateSkinRestrictionWindow(ctx iris.Context) {
	windowID, err := parsePathUint32(ctx.Params().Get("id"), "window id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.SkinRestrictionWindowUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	var window orm.GlobalSkinRestrictionWindow
	if err := orm.GormDB.First(&window, windowID).Error; err != nil {
		writeGameDataError(ctx, err, "skin restriction window")
		return
	}

	window.SkinID = req.SkinID
	window.Type = req.Type
	window.StartTime = req.StartTime
	window.StopTime = req.StopTime

	if err := orm.GormDB.Save(&window).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update skin restriction window", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteSkinRestrictionWindow godoc
// @Summary     Delete skin restriction window
// @Tags        SkinRestrictionWindows
// @Produce     json
// @Param       id   path      int  true  "Window ID"
// @Success     200  {object}  SkinRestrictionWindowMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/skin-restriction-windows/{id} [delete]
func (handler *GameDataHandler) DeleteSkinRestrictionWindow(ctx iris.Context) {
	windowID, err := parsePathUint32(ctx.Params().Get("id"), "window id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.GlobalSkinRestrictionWindow{}, windowID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete skin restriction window", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "skin restriction window not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ShipSkins godoc
// @Summary     List ship skins
// @Tags        Ships
// @Produce     json
// @Param       id      path   int  true  "Ship ID"
// @Param       offset  query  int  false "Pagination offset"
// @Param       limit   query  int  false "Pagination limit"
// @Success     200  {object}  ListSkinsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/ships/{id}/skins [get]
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

func skinPayloadFromModel(skin orm.Skin) types.SkinPayload {
	return types.SkinPayload{
		ID:             skin.ID,
		Name:           skin.Name,
		ShipGroup:      skin.ShipGroup,
		Desc:           skin.Desc,
		BG:             skin.BG,
		BGSp:           skin.BGSp,
		BGM:            skin.BGM,
		Painting:       skin.Painting,
		Prefab:         skin.Prefab,
		ChangeSkin:     types.RawJSON{Value: skin.ChangeSkin},
		ShowSkin:       skin.ShowSkin,
		SkeletonSkin:   skin.SkeletonSkin,
		ShipL2DID:      types.RawJSON{Value: skin.ShipL2DID},
		L2DAnimations:  types.RawJSON{Value: skin.L2DAnimations},
		L2DDragRate:    types.RawJSON{Value: skin.L2DDragRate},
		L2DParaRange:   types.RawJSON{Value: skin.L2DParaRange},
		L2DSE:          types.RawJSON{Value: skin.L2DSE},
		L2DVoiceCalib:  types.RawJSON{Value: skin.L2DVoiceCalib},
		PartScale:      skin.PartScale,
		MainUIFX:       skin.MainUIFX,
		SpineOffset:    types.RawJSON{Value: skin.SpineOffset},
		SpineProfile:   types.RawJSON{Value: skin.SpineProfile},
		Tag:            types.RawJSON{Value: skin.Tag},
		Time:           types.RawJSON{Value: skin.Time},
		GetShowing:     types.RawJSON{Value: skin.GetShowing},
		PurchaseOffset: types.RawJSON{Value: skin.PurchaseOffset},
		ShopOffset:     types.RawJSON{Value: skin.ShopOffset},
		RarityBG:       skin.RarityBG,
		SpecialEffects: types.RawJSON{Value: skin.SpecialEffects},
		GroupIndex:     skin.GroupIndex,
		Gyro:           skin.Gyro,
		HandID:         skin.HandID,
		Illustrator:    skin.Illustrator,
		Illustrator2:   skin.Illustrator2,
		VoiceActor:     skin.VoiceActor,
		VoiceActor2:    skin.VoiceActor2,
		DoubleChar:     skin.DoubleChar,
		LipSmoothing:   skin.LipSmoothing,
		LipSyncGain:    skin.LipSyncGain,
		L2DIgnoreDrag:  skin.L2DIgnoreDrag,
		SkinType:       skin.SkinType,
		ShopID:         skin.ShopID,
		ShopTypeID:     skin.ShopTypeID,
		ShopDynamicHX:  skin.ShopDynamicHX,
		SpineAction:    types.RawJSON{Value: skin.SpineAction},
		SpineUseLive2D: skin.SpineUseLive2D,
		Live2DOffset:   types.RawJSON{Value: skin.Live2DOffset},
		Live2DProfile:  types.RawJSON{Value: skin.Live2DProfile},
		FXContainer:    types.RawJSON{Value: skin.FXContainer},
		BoundBone:      types.RawJSON{Value: skin.BoundBone},
		Smoke:          types.RawJSON{Value: skin.Smoke},
	}
}

func applySkinPayload(skin *orm.Skin, payload types.SkinPayload) {
	skin.Name = payload.Name
	skin.ShipGroup = payload.ShipGroup
	skin.Desc = payload.Desc
	skin.BG = payload.BG
	skin.BGSp = payload.BGSp
	skin.BGM = payload.BGM
	skin.Painting = payload.Painting
	skin.Prefab = payload.Prefab
	skin.ChangeSkin = payload.ChangeSkin.Value
	skin.ShowSkin = payload.ShowSkin
	skin.SkeletonSkin = payload.SkeletonSkin
	skin.ShipL2DID = payload.ShipL2DID.Value
	skin.L2DAnimations = payload.L2DAnimations.Value
	skin.L2DDragRate = payload.L2DDragRate.Value
	skin.L2DParaRange = payload.L2DParaRange.Value
	skin.L2DSE = payload.L2DSE.Value
	skin.L2DVoiceCalib = payload.L2DVoiceCalib.Value
	skin.PartScale = payload.PartScale
	skin.MainUIFX = payload.MainUIFX
	skin.SpineOffset = payload.SpineOffset.Value
	skin.SpineProfile = payload.SpineProfile.Value
	skin.Tag = payload.Tag.Value
	skin.Time = payload.Time.Value
	skin.GetShowing = payload.GetShowing.Value
	skin.PurchaseOffset = payload.PurchaseOffset.Value
	skin.ShopOffset = payload.ShopOffset.Value
	skin.RarityBG = payload.RarityBG
	skin.SpecialEffects = payload.SpecialEffects.Value
	skin.GroupIndex = payload.GroupIndex
	skin.Gyro = payload.Gyro
	skin.HandID = payload.HandID
	skin.Illustrator = payload.Illustrator
	skin.Illustrator2 = payload.Illustrator2
	skin.VoiceActor = payload.VoiceActor
	skin.VoiceActor2 = payload.VoiceActor2
	skin.DoubleChar = payload.DoubleChar
	skin.LipSmoothing = payload.LipSmoothing
	skin.LipSyncGain = payload.LipSyncGain
	skin.L2DIgnoreDrag = payload.L2DIgnoreDrag
	skin.SkinType = payload.SkinType
	skin.ShopID = payload.ShopID
	skin.ShopTypeID = payload.ShopTypeID
	skin.ShopDynamicHX = payload.ShopDynamicHX
	skin.SpineAction = payload.SpineAction.Value
	skin.SpineUseLive2D = payload.SpineUseLive2D
	skin.Live2DOffset = payload.Live2DOffset.Value
	skin.Live2DProfile = payload.Live2DProfile.Value
	skin.FXContainer = payload.FXContainer.Value
	skin.BoundBone = payload.BoundBone.Value
	skin.Smoke = payload.Smoke.Value
}

func equipmentPayloadFromModel(equipment orm.Equipment) types.EquipmentPayload {
	return types.EquipmentPayload{
		ID:                equipment.ID,
		Base:              equipment.Base,
		DestroyGold:       equipment.DestroyGold,
		DestroyItem:       types.RawJSON{Value: equipment.DestroyItem},
		EquipLimit:        equipment.EquipLimit,
		Group:             equipment.Group,
		Important:         equipment.Important,
		Level:             equipment.Level,
		Next:              equipment.Next,
		Prev:              equipment.Prev,
		RestoreGold:       equipment.RestoreGold,
		RestoreItem:       types.RawJSON{Value: equipment.RestoreItem},
		ShipTypeForbidden: types.RawJSON{Value: equipment.ShipTypeForbidden},
		TransUseGold:      equipment.TransUseGold,
		TransUseItem:      types.RawJSON{Value: equipment.TransUseItem},
		Type:              equipment.Type,
		UpgradeFormulaID:  types.RawJSON{Value: equipment.UpgradeFormulaID},
	}
}

func applyEquipmentPayload(equipment *orm.Equipment, payload types.EquipmentPayload) {
	equipment.Base = payload.Base
	equipment.DestroyGold = payload.DestroyGold
	equipment.DestroyItem = payload.DestroyItem.Value
	equipment.EquipLimit = payload.EquipLimit
	equipment.Group = payload.Group
	equipment.Important = payload.Important
	equipment.Level = payload.Level
	equipment.Next = payload.Next
	equipment.Prev = payload.Prev
	equipment.RestoreGold = payload.RestoreGold
	equipment.RestoreItem = payload.RestoreItem.Value
	equipment.ShipTypeForbidden = payload.ShipTypeForbidden.Value
	equipment.TransUseGold = payload.TransUseGold
	equipment.TransUseItem = payload.TransUseItem.Value
	equipment.Type = payload.Type
	equipment.UpgradeFormulaID = payload.UpgradeFormulaID.Value
}

func weaponPayloadFromModel(weapon orm.Weapon) types.WeaponPayload {
	return types.WeaponPayload{
		ID:                   weapon.ID,
		ActionIndex:          weapon.ActionIndex,
		AimType:              weapon.AimType,
		Angle:                weapon.Angle,
		AttackAttribute:      weapon.AttackAttribute,
		AttackAttributeRatio: weapon.AttackAttributeRatio,
		AutoAftercast:        types.RawJSON{Value: weapon.AutoAftercast},
		AxisAngle:            weapon.AxisAngle,
		BarrageID:            types.RawJSON{Value: weapon.BarrageID},
		BulletID:             types.RawJSON{Value: weapon.BulletID},
		ChargeParam:          types.RawJSON{Value: weapon.ChargeParam},
		Corrected:            weapon.Corrected,
		Damage:               weapon.Damage,
		EffectMove:           weapon.EffectMove,
		Expose:               weapon.Expose,
		FireFX:               weapon.FireFX,
		FireFXLoopType:       weapon.FireFXLoopType,
		FireSFX:              weapon.FireSFX,
		InitialOverHeat:      weapon.InitialOverHeat,
		MinRange:             weapon.MinRange,
		OxyType:              types.RawJSON{Value: weapon.OxyType},
		PrecastParam:         types.RawJSON{Value: weapon.PrecastParam},
		Queue:                weapon.Queue,
		Range:                weapon.Range,
		RecoverTime:          types.RawJSON{Value: weapon.RecoverTime},
		ReloadMax:            weapon.ReloadMax,
		SearchCondition:      types.RawJSON{Value: weapon.SearchCondition},
		SearchType:           weapon.SearchType,
		ShakeScreen:          weapon.ShakeScreen,
		SpawnBound:           types.RawJSON{Value: weapon.SpawnBound},
		Suppress:             weapon.Suppress,
		TorpedoAmmo:          weapon.TorpedoAmmo,
		Type:                 weapon.Type,
	}
}

func applyWeaponPayload(weapon *orm.Weapon, payload types.WeaponPayload) {
	weapon.ActionIndex = payload.ActionIndex
	weapon.AimType = payload.AimType
	weapon.Angle = payload.Angle
	weapon.AttackAttribute = payload.AttackAttribute
	weapon.AttackAttributeRatio = payload.AttackAttributeRatio
	weapon.AutoAftercast = payload.AutoAftercast.Value
	weapon.AxisAngle = payload.AxisAngle
	weapon.BarrageID = payload.BarrageID.Value
	weapon.BulletID = payload.BulletID.Value
	weapon.ChargeParam = payload.ChargeParam.Value
	weapon.Corrected = payload.Corrected
	weapon.Damage = payload.Damage
	weapon.EffectMove = payload.EffectMove
	weapon.Expose = payload.Expose
	weapon.FireFX = payload.FireFX
	weapon.FireFXLoopType = payload.FireFXLoopType
	weapon.FireSFX = payload.FireSFX
	weapon.InitialOverHeat = payload.InitialOverHeat
	weapon.MinRange = payload.MinRange
	weapon.OxyType = payload.OxyType.Value
	weapon.PrecastParam = payload.PrecastParam.Value
	weapon.Queue = payload.Queue
	weapon.Range = payload.Range
	weapon.RecoverTime = payload.RecoverTime.Value
	weapon.ReloadMax = payload.ReloadMax
	weapon.SearchCondition = payload.SearchCondition.Value
	weapon.SearchType = payload.SearchType
	weapon.ShakeScreen = payload.ShakeScreen
	weapon.SpawnBound = payload.SpawnBound.Value
	weapon.Suppress = payload.Suppress
	weapon.TorpedoAmmo = payload.TorpedoAmmo
	weapon.Type = payload.Type
}

func skillPayloadFromModel(skill orm.Skill) types.SkillPayload {
	return types.SkillPayload{
		ID:         skill.ID,
		Name:       skill.Name,
		Desc:       skill.Desc,
		CD:         skill.CD,
		Painting:   types.RawJSON{Value: skill.Painting},
		Picture:    skill.Picture,
		AniEffect:  types.RawJSON{Value: skill.AniEffect},
		UIEffect:   skill.UIEffect,
		EffectList: types.RawJSON{Value: skill.EffectList},
	}
}

func applySkillPayload(skill *orm.Skill, payload types.SkillPayload) {
	skill.Desc = payload.Desc
	skill.CD = payload.CD
	skill.Painting = payload.Painting.Value
	skill.Picture = payload.Picture
	skill.AniEffect = payload.AniEffect.Value
	skill.UIEffect = payload.UIEffect
	skill.EffectList = payload.EffectList.Value
}

func buffPayloadFromModel(buff orm.Buff) types.BuffPayload {
	return types.BuffPayload{
		ID:          buff.ID,
		Name:        buff.Name,
		Desc:        buff.Description,
		MaxTime:     buff.MaxTime,
		BenefitType: buff.BenefitType,
	}
}

// ListConfigEntries godoc
// @Summary     List config entries
// @Tags        ConfigEntries
// @Produce     json
// @Param       category  query     string  false  "Filter by category"
// @Param       key       query     string  false  "Filter by key"
// @Success     200  {object}  ConfigEntryListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/config-entries [get]
func (handler *GameDataHandler) ListConfigEntries(ctx iris.Context) {
	category := strings.TrimSpace(ctx.URLParam("category"))
	key := strings.TrimSpace(ctx.URLParam("key"))

	query := orm.GormDB.Model(&orm.ConfigEntry{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if key != "" {
		query = query.Where("key = ?", key)
	}

	var entries []orm.ConfigEntry
	if err := query.Order("category asc").Order("key asc").Find(&entries).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list config entries", nil))
		return
	}

	payload := types.ConfigEntryListResponse{Entries: make([]types.ConfigEntryPayload, 0, len(entries))}
	for _, entry := range entries {
		payload.Entries = append(payload.Entries, configEntryPayloadFromModel(entry))
	}

	_ = ctx.JSON(response.Success(payload))
}

// ConfigEntryDetail godoc
// @Summary     Get config entry
// @Tags        ConfigEntries
// @Produce     json
// @Param       id   path      int  true  "Config entry ID"
// @Success     200  {object}  ConfigEntryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/config-entries/{id} [get]
func (handler *GameDataHandler) ConfigEntryDetail(ctx iris.Context) {
	entryID, err := parsePathUint64(ctx.Params().Get("id"), "config entry id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var entry orm.ConfigEntry
	if err := orm.GormDB.First(&entry, "id = ?", entryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "config entry not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load config entry", nil))
		return
	}

	_ = ctx.JSON(response.Success(configEntryPayloadFromModel(entry)))
}

// CreateConfigEntry godoc
// @Summary     Create config entry
// @Tags        ConfigEntries
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ConfigEntryMutationRequest  true  "Config entry"
// @Success     200  {object}  ConfigEntryMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/config-entries [post]
func (handler *GameDataHandler) CreateConfigEntry(ctx iris.Context) {
	var req types.ConfigEntryMutationRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	category := strings.TrimSpace(req.Category)
	key := strings.TrimSpace(req.Key)
	if category == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "category is required", nil))
		return
	}
	if key == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "key is required", nil))
		return
	}
	if req.Data.Value == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "data is required", nil))
		return
	}

	entry := orm.ConfigEntry{
		Category: category,
		Key:      key,
		Data:     req.Data.Value,
	}

	if err := orm.GormDB.Create(&entry).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create config entry", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateConfigEntry godoc
// @Summary     Update config entry
// @Tags        ConfigEntries
// @Accept      json
// @Produce     json
// @Param       id       path      int                                true  "Config entry ID"
// @Param       payload  body      types.ConfigEntryMutationRequest  true  "Config entry"
// @Success     200  {object}  ConfigEntryMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/config-entries/{id} [put]
func (handler *GameDataHandler) UpdateConfigEntry(ctx iris.Context) {
	entryID, err := parsePathUint64(ctx.Params().Get("id"), "config entry id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.ConfigEntryMutationRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	category := strings.TrimSpace(req.Category)
	key := strings.TrimSpace(req.Key)
	if category == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "category is required", nil))
		return
	}
	if key == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "key is required", nil))
		return
	}
	if req.Data.Value == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "data is required", nil))
		return
	}

	var entry orm.ConfigEntry
	if err := orm.GormDB.First(&entry, "id = ?", entryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "config entry not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load config entry", nil))
		return
	}

	entry.Category = category
	entry.Key = key
	entry.Data = req.Data.Value

	if err := orm.GormDB.Save(&entry).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update config entry", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteConfigEntry godoc
// @Summary     Delete config entry
// @Tags        ConfigEntries
// @Produce     json
// @Param       id   path      int  true  "Config entry ID"
// @Success     200  {object}  ConfigEntryMutationResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/config-entries/{id} [delete]
func (handler *GameDataHandler) DeleteConfigEntry(ctx iris.Context) {
	entryID, err := parsePathUint64(ctx.Params().Get("id"), "config entry id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.ConfigEntry{}, "id = ?", entryID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete config entry", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "config entry not found", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListLivingAreaCovers godoc
// @Summary     List living area covers
// @Tags        GameData
// @Produce     json
// @Success     200  {object}  ConfigEntryListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/livingarea-covers [get]
func (handler *GameDataHandler) ListLivingAreaCovers(ctx iris.Context) {
	listConfigEntries(ctx, "ShareCfg/livingarea_cover.json")
}

// ListIconFrames godoc
// @Summary     List icon frames
// @Tags        GameData
// @Produce     json
// @Success     200  {object}  ConfigEntryListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/attire/icon-frames [get]
func (handler *GameDataHandler) ListIconFrames(ctx iris.Context) {
	listConfigEntries(ctx, "ShareCfg/item_data_frame.json")
}

// ListChatFrames godoc
// @Summary     List chat frames
// @Tags        GameData
// @Produce     json
// @Success     200  {object}  ConfigEntryListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/attire/chat-frames [get]
func (handler *GameDataHandler) ListChatFrames(ctx iris.Context) {
	listConfigEntries(ctx, "ShareCfg/item_data_chat.json")
}

// ListBattleUIStyles godoc
// @Summary     List battle UI styles
// @Tags        GameData
// @Produce     json
// @Success     200  {object}  ConfigEntryListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/attire/battle-ui [get]
func (handler *GameDataHandler) ListBattleUIStyles(ctx iris.Context) {
	listConfigEntries(ctx, "ShareCfg/item_data_battleui.json")
}

func listConfigEntries(ctx iris.Context, category string) {
	entries, err := orm.ListConfigEntries(orm.GormDB, category)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load config entries", nil))
		return
	}
	payload := types.ConfigEntryListResponse{Entries: make([]types.ConfigEntryPayload, 0, len(entries))}
	for _, entry := range entries {
		payload.Entries = append(payload.Entries, configEntryPayloadFromModel(entry))
	}
	_ = ctx.JSON(response.Success(payload))
}

func configEntryPayloadFromModel(entry orm.ConfigEntry) types.ConfigEntryPayload {
	return types.ConfigEntryPayload{
		ID:       entry.ID,
		Category: entry.Category,
		Key:      entry.Key,
		Data:     types.RawJSON{Value: entry.Data},
	}
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

func parsePathUint64(value string, name string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return parsed, nil
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
