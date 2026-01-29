package handlers

import (
	"errors"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type Dorm3dHandler struct{}

func NewDorm3dHandler() *Dorm3dHandler {
	return &Dorm3dHandler{}
}

func RegisterDorm3dRoutes(party iris.Party, handler *Dorm3dHandler) {
	party.Get("", handler.ListDorm3dApartments)
	party.Get("/{id:uint}", handler.Dorm3dApartmentDetail)
	party.Get("/{id:uint}/gifts", handler.Dorm3dApartmentGifts)
	party.Put("/{id:uint}/gifts", handler.UpdateDorm3dApartmentGifts)
	party.Get("/{id:uint}/ships", handler.Dorm3dApartmentShips)
	party.Put("/{id:uint}/ships", handler.UpdateDorm3dApartmentShips)
	party.Get("/{id:uint}/rooms", handler.Dorm3dApartmentRooms)
	party.Put("/{id:uint}/rooms", handler.UpdateDorm3dApartmentRooms)
	party.Get("/{id:uint}/ins", handler.Dorm3dApartmentIns)
	party.Put("/{id:uint}/ins", handler.UpdateDorm3dApartmentIns)
	party.Post("", handler.CreateDorm3dApartment)
	party.Put("/{id:uint}", handler.UpdateDorm3dApartment)
	party.Delete("/{id:uint}", handler.DeleteDorm3dApartment)
}

// ListDorm3dApartments godoc
// @Summary     List Dorm3d apartments
// @Tags        Dorm3d
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  Dorm3dApartmentListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments [get]
func (handler *Dorm3dHandler) ListDorm3dApartments(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var total int64
	if err := orm.GormDB.Model(&orm.Dorm3dApartment{}).Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list dorm3d apartments", nil))
		return
	}

	var apartments []orm.Dorm3dApartment
	query := orm.GormDB.Order("commander_id asc")
	query = orm.ApplyPagination(query, pagination.Offset, pagination.Limit)
	if err := query.Find(&apartments).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list dorm3d apartments", nil))
		return
	}

	payload := types.Dorm3dApartmentListResponse{
		Apartments: apartments,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// Dorm3dApartmentDetail godoc
// @Summary     Get Dorm3d apartment
// @Tags        Dorm3d
// @Produce     json
// @Param       id   path      int  true  "Commander ID"
// @Success     200  {object}  Dorm3dApartmentResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id} [get]
func (handler *Dorm3dHandler) Dorm3dApartmentDetail(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var apartment orm.Dorm3dApartment
	if err := orm.GormDB.First(&apartment, "commander_id = ?", commanderID).Error; err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}
	apartment.EnsureDefaults()
	_ = ctx.JSON(response.Success(apartment))
}

// Dorm3dApartmentGifts godoc
// @Summary     Get Dorm3d apartment gifts
// @Tags        Dorm3d
// @Produce     json
// @Param       id   path      int  true  "Commander ID"
// @Success     200  {object}  Dorm3dApartmentGiftsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/gifts [get]
func (handler *Dorm3dHandler) Dorm3dApartmentGifts(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	_ = ctx.JSON(response.Success(apartment.Gifts))
}

// UpdateDorm3dApartmentGifts godoc
// @Summary     Update Dorm3d apartment gifts
// @Tags        Dorm3d
// @Accept      json
// @Produce     json
// @Param       id       path      int                   true  "Commander ID"
// @Param       payload  body      types.Dorm3dGiftList  true  "Dorm3d gifts"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/gifts [put]
func (handler *Dorm3dHandler) UpdateDorm3dApartmentGifts(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.Dorm3dGiftList
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	apartment.Gifts = req
	if err := orm.SaveDorm3dApartment(apartment); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update dorm3d gifts", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// Dorm3dApartmentShips godoc
// @Summary     Get Dorm3d apartment ships
// @Tags        Dorm3d
// @Produce     json
// @Param       id   path      int  true  "Commander ID"
// @Success     200  {object}  Dorm3dApartmentShipsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/ships [get]
func (handler *Dorm3dHandler) Dorm3dApartmentShips(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	_ = ctx.JSON(response.Success(apartment.Ships))
}

// UpdateDorm3dApartmentShips godoc
// @Summary     Update Dorm3d apartment ships
// @Tags        Dorm3d
// @Accept      json
// @Produce     json
// @Param       id       path      int                   true  "Commander ID"
// @Param       payload  body      types.Dorm3dShipList  true  "Dorm3d ships"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/ships [put]
func (handler *Dorm3dHandler) UpdateDorm3dApartmentShips(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.Dorm3dShipList
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	apartment.Ships = req
	if err := orm.SaveDorm3dApartment(apartment); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update dorm3d ships", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// Dorm3dApartmentRooms godoc
// @Summary     Get Dorm3d apartment rooms
// @Tags        Dorm3d
// @Produce     json
// @Param       id   path      int  true  "Commander ID"
// @Success     200  {object}  Dorm3dApartmentRoomsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/rooms [get]
func (handler *Dorm3dHandler) Dorm3dApartmentRooms(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	_ = ctx.JSON(response.Success(apartment.Rooms))
}

// UpdateDorm3dApartmentRooms godoc
// @Summary     Update Dorm3d apartment rooms
// @Tags        Dorm3d
// @Accept      json
// @Produce     json
// @Param       id       path      int                   true  "Commander ID"
// @Param       payload  body      types.Dorm3dRoomList  true  "Dorm3d rooms"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/rooms [put]
func (handler *Dorm3dHandler) UpdateDorm3dApartmentRooms(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.Dorm3dRoomList
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	apartment.Rooms = req
	if err := orm.SaveDorm3dApartment(apartment); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update dorm3d rooms", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// Dorm3dApartmentIns godoc
// @Summary     Get Dorm3d apartment ins
// @Tags        Dorm3d
// @Produce     json
// @Param       id   path      int  true  "Commander ID"
// @Success     200  {object}  Dorm3dApartmentInsResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/ins [get]
func (handler *Dorm3dHandler) Dorm3dApartmentIns(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	_ = ctx.JSON(response.Success(apartment.Ins))
}

// UpdateDorm3dApartmentIns godoc
// @Summary     Update Dorm3d apartment ins
// @Tags        Dorm3d
// @Accept      json
// @Produce     json
// @Param       id       path      int                  true  "Commander ID"
// @Param       payload  body      types.Dorm3dInsList  true  "Dorm3d ins"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id}/ins [put]
func (handler *Dorm3dHandler) UpdateDorm3dApartmentIns(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.Dorm3dInsList
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	apartment, err := orm.GetDorm3dApartment(commanderID)
	if err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	apartment.Ins = req
	if err := orm.SaveDorm3dApartment(apartment); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update dorm3d ins", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// CreateDorm3dApartment godoc
// @Summary     Create Dorm3d apartment
// @Tags        Dorm3d
// @Accept      json
// @Produce     json
// @Param       payload  body      types.Dorm3dApartmentRequest  true  "Dorm3d apartment"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments [post]
func (handler *Dorm3dHandler) CreateDorm3dApartment(ctx iris.Context) {
	req, err := readDorm3dApartmentRequest(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if req.CommanderID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "commander_id is required", nil))
		return
	}
	req.EnsureDefaults()
	if err := orm.GormDB.Create(req).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create dorm3d apartment", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdateDorm3dApartment godoc
// @Summary     Update Dorm3d apartment
// @Tags        Dorm3d
// @Accept      json
// @Produce     json
// @Param       id       path      int                         true  "Commander ID"
// @Param       payload  body      types.Dorm3dApartmentRequest  true  "Dorm3d apartment"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id} [put]
func (handler *Dorm3dHandler) UpdateDorm3dApartment(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	req, err := readDorm3dApartmentRequest(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var apartment orm.Dorm3dApartment
	if err := orm.GormDB.First(&apartment, "commander_id = ?", commanderID).Error; err != nil {
		writeDorm3dApartmentError(ctx, err)
		return
	}

	req.CommanderID = commanderID
	req.EnsureDefaults()
	apartment = *req
	if err := orm.GormDB.Save(&apartment).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update dorm3d apartment", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeleteDorm3dApartment godoc
// @Summary     Delete Dorm3d apartment
// @Tags        Dorm3d
// @Produce     json
// @Param       id   path      int  true  "Commander ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/dorm3d-apartments/{id} [delete]
func (handler *Dorm3dHandler) DeleteDorm3dApartment(ctx iris.Context) {
	commanderID, err := parsePathUint32(ctx.Params().Get("id"), "commander id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result := orm.GormDB.Delete(&orm.Dorm3dApartment{}, "commander_id = ?", commanderID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete dorm3d apartment", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "dorm3d apartment not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

func readDorm3dApartmentRequest(ctx iris.Context) (*types.Dorm3dApartmentRequest, error) {
	var req types.Dorm3dApartmentRequest
	if err := ctx.ReadJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func writeDorm3dApartmentError(ctx iris.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "dorm3d apartment not found", nil))
		return
	}
	ctx.StatusCode(iris.StatusInternalServerError)
	_ = ctx.JSON(response.Error("internal_error", "failed to fetch dorm3d apartment", nil))
}
