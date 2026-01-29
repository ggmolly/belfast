package handlers

import (
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type ShopHandler struct{}

type NoticeHandler struct{}

func NewShopHandler() *ShopHandler {
	return &ShopHandler{}
}

func NewNoticeHandler() *NoticeHandler {
	return &NoticeHandler{}
}

func RegisterShopRoutes(party iris.Party, handler *ShopHandler) {
	party.Get("/offers", handler.ListOffers)
	party.Post("/offers", handler.CreateOffer)
	party.Put("/offers/{id:uint}", handler.UpdateOffer)
	party.Delete("/offers/{id:uint}", handler.DeleteOffer)
}

func RegisterNoticeRoutes(party iris.Party, handler *NoticeHandler) {
	party.Get("", handler.ListNotices)
	party.Get("/{id:uint}", handler.GetNotice)
	party.Post("", handler.CreateNotice)
	party.Put("/{id:uint}", handler.UpdateNotice)
	party.Delete("/{id:uint}", handler.DeleteNotice)
	party.Get("/active", handler.ActiveNotices)
}

// ListOffers godoc
// @Summary     List shop offers
// @Tags        Shop
// @Produce     json
// @Param       offset  query  int     false  "Pagination offset"
// @Param       limit   query  int     false  "Pagination limit"
// @Param       genre   query  string  false  "Filter by genre"
// @Success     200  {object}  ShopOfferListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/shop/offers [get]
func (handler *ShopHandler) ListOffers(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	genre := strings.TrimSpace(ctx.URLParam("genre"))
	result, err := orm.ListShopOffers(orm.GormDB, orm.ShopOfferQueryParams{Offset: pagination.Offset, Limit: pagination.Limit, Genre: genre})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list shop offers", nil))
		return
	}

	offers := make([]types.ShopOfferSummary, 0, len(result.Offers))
	for _, offer := range result.Offers {
		offers = append(offers, types.ShopOfferSummary{
			ID:             offer.ID,
			Effects:        []int64(offer.Effects),
			EffectArgs:     types.RawJSON{Value: offer.EffectArgs},
			Number:         offer.Number,
			ResourceNumber: offer.ResourceNumber,
			ResourceID:     offer.ResourceID,
			Type:           offer.Type,
			Genre:          offer.Genre,
			Discount:       offer.Discount,
		})
	}

	payload := types.ShopOfferListResponse{
		Offers: offers,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  result.Total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateOffer godoc
// @Summary     Create shop offer
// @Tags        Shop
// @Accept      json
// @Produce     json
// @Param       offer  body      types.ShopOfferSummary  true  "Shop offer"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/shop/offers [post]
func (handler *ShopHandler) CreateOffer(ctx iris.Context) {
	var req types.ShopOfferSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	offer := orm.ShopOffer{
		ID:             req.ID,
		Effects:        orm.Int64List(req.Effects),
		EffectArgs:     req.EffectArgs.Value,
		Number:         req.Number,
		ResourceNumber: req.ResourceNumber,
		ResourceID:     req.ResourceID,
		Type:           req.Type,
		Genre:          req.Genre,
		Discount:       req.Discount,
	}

	if err := orm.GormDB.Create(&offer).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create shop offer", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateOffer godoc
// @Summary     Update shop offer
// @Tags        Shop
// @Accept      json
// @Produce     json
// @Param       id     path      int                    true  "Offer ID"
// @Param       offer  body      types.ShopOfferSummary  true  "Shop offer"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/shop/offers/{id} [put]
func (handler *ShopHandler) UpdateOffer(ctx iris.Context) {
	offerID, err := parsePathUint32(ctx.Params().Get("id"), "offer id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.ShopOfferSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	var offer orm.ShopOffer
	if err := orm.GormDB.First(&offer, offerID).Error; err != nil {
		writeShopNoticeError(ctx, err, "shop offer")
		return
	}

	offer.Effects = orm.Int64List(req.Effects)
	offer.EffectArgs = req.EffectArgs.Value
	offer.Number = req.Number
	offer.ResourceNumber = req.ResourceNumber
	offer.ResourceID = req.ResourceID
	offer.Type = req.Type
	offer.Genre = req.Genre
	offer.Discount = req.Discount

	if err := orm.GormDB.Save(&offer).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update shop offer", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteOffer godoc
// @Summary     Delete shop offer
// @Tags        Shop
// @Produce     json
// @Param       id   path  int  true  "Offer ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/shop/offers/{id} [delete]
func (handler *ShopHandler) DeleteOffer(ctx iris.Context) {
	offerID, err := parsePathUint32(ctx.Params().Get("id"), "offer id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if err := orm.GormDB.Delete(&orm.ShopOffer{}, offerID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete shop offer", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ListNotices godoc
// @Summary     List notices
// @Tags        Notices
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  NoticeListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/notices [get]
func (handler *NoticeHandler) ListNotices(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result, err := orm.ListNotices(orm.GormDB, orm.NoticeQueryParams{Offset: pagination.Offset, Limit: pagination.Limit})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list notices", nil))
		return
	}

	notices := make([]types.NoticeSummary, 0, len(result.Notices))
	for _, notice := range result.Notices {
		notices = append(notices, types.NoticeSummary{
			ID:         notice.ID,
			Version:    notice.Version,
			BtnTitle:   notice.BtnTitle,
			Title:      notice.Title,
			TitleImage: notice.TitleImage,
			TimeDesc:   notice.TimeDesc,
			Content:    notice.Content,
			TagType:    notice.TagType,
			Icon:       notice.Icon,
			Track:      notice.Track,
		})
	}

	payload := types.NoticeListResponse{
		Notices: notices,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  result.Total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// GetNotice godoc
// @Summary     Get notice
// @Tags        Notices
// @Produce     json
// @Param       id   path  int  true  "Notice ID"
// @Success     200  {object}  NoticeSummaryResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/notices/{id} [get]
func (handler *NoticeHandler) GetNotice(ctx iris.Context) {
	noticeID, err := parsePathUint32(ctx.Params().Get("id"), "notice id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var notice orm.Notice
	if err := orm.GormDB.First(&notice, noticeID).Error; err != nil {
		writeShopNoticeError(ctx, err, "notice")
		return
	}

	payload := types.NoticeSummary{
		ID:         notice.ID,
		Version:    notice.Version,
		BtnTitle:   notice.BtnTitle,
		Title:      notice.Title,
		TitleImage: notice.TitleImage,
		TimeDesc:   notice.TimeDesc,
		Content:    notice.Content,
		TagType:    notice.TagType,
		Icon:       notice.Icon,
		Track:      notice.Track,
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateNotice godoc
// @Summary     Create notice
// @Tags        Notices
// @Accept      json
// @Produce     json
// @Param       notice  body      types.NoticeSummary  true  "Notice"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/notices [post]
func (handler *NoticeHandler) CreateNotice(ctx iris.Context) {
	var req types.NoticeSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "title and content are required", nil))
		return
	}

	notice := orm.Notice{
		ID:         req.ID,
		Version:    req.Version,
		BtnTitle:   req.BtnTitle,
		Title:      req.Title,
		TitleImage: req.TitleImage,
		TimeDesc:   req.TimeDesc,
		Content:    req.Content,
		TagType:    req.TagType,
		Icon:       req.Icon,
		Track:      req.Track,
	}

	if err := orm.GormDB.Create(&notice).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create notice", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateNotice godoc
// @Summary     Update notice
// @Tags        Notices
// @Accept      json
// @Produce     json
// @Param       id      path  int                 true  "Notice ID"
// @Param       notice  body  types.NoticeSummary  true  "Notice"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/notices/{id} [put]
func (handler *NoticeHandler) UpdateNotice(ctx iris.Context) {
	noticeID, err := parsePathUint32(ctx.Params().Get("id"), "notice id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var req types.NoticeSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	var notice orm.Notice
	if err := orm.GormDB.First(&notice, noticeID).Error; err != nil {
		writeShopNoticeError(ctx, err, "notice")
		return
	}

	notice.Version = req.Version
	notice.BtnTitle = req.BtnTitle
	notice.Title = req.Title
	notice.TitleImage = req.TitleImage
	notice.TimeDesc = req.TimeDesc
	notice.Content = req.Content
	notice.TagType = req.TagType
	notice.Icon = req.Icon
	notice.Track = req.Track

	if err := orm.GormDB.Save(&notice).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update notice", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteNotice godoc
// @Summary     Delete notice
// @Tags        Notices
// @Produce     json
// @Param       id   path  int  true  "Notice ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/notices/{id} [delete]
func (handler *NoticeHandler) DeleteNotice(ctx iris.Context) {
	noticeID, err := parsePathUint32(ctx.Params().Get("id"), "notice id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	if err := orm.GormDB.Delete(&orm.Notice{}, noticeID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete notice", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// ActiveNotices godoc
// @Summary     List active notices
// @Tags        Notices
// @Produce     json
// @Success     200  {object}  NoticeActiveResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/notices/active [get]
func (handler *NoticeHandler) ActiveNotices(ctx iris.Context) {
	notices, err := orm.ListActiveNotices(orm.GormDB)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load active notices", nil))
		return
	}

	payload := make([]types.NoticeSummary, 0, len(notices))
	for _, notice := range notices {
		payload = append(payload, types.NoticeSummary{
			ID:         notice.ID,
			Version:    notice.Version,
			BtnTitle:   notice.BtnTitle,
			Title:      notice.Title,
			TitleImage: notice.TitleImage,
			TimeDesc:   notice.TimeDesc,
			Content:    notice.Content,
			TagType:    notice.TagType,
			Icon:       notice.Icon,
			Track:      notice.Track,
		})
	}

	_ = ctx.JSON(response.Success(payload))
}

func writeShopNoticeError(ctx iris.Context, err error, item string) {
	if err == gorm.ErrRecordNotFound {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", item+" not found", nil))
		return
	}
	ctx.StatusCode(iris.StatusInternalServerError)
	_ = ctx.JSON(response.Error("internal_error", "failed to load "+item, nil))
}
