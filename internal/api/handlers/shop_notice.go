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
	party.Post("", handler.CreateNotice)
	party.Put("/{id:uint}", handler.UpdateNotice)
	party.Delete("/{id:uint}", handler.DeleteNotice)
	party.Get("/active", handler.ActiveNotices)
}

func (handler *ShopHandler) ListOffers(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	result, err := orm.ListShopOffers(orm.GormDB, orm.ShopOfferQueryParams{Offset: pagination.Offset, Limit: pagination.Limit})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list shop offers", nil))
		return
	}

	offers := make([]types.ShopOfferSummary, 0, len(result.Offers))
	for _, offer := range result.Offers {
		offers = append(offers, types.ShopOfferSummary{
			ID:             offer.ID,
			Effects:        offer.Effects,
			Number:         offer.Number,
			ResourceNumber: offer.ResourceNumber,
			ResourceID:     offer.ResourceID,
			Type:           offer.Type,
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

func (handler *ShopHandler) CreateOffer(ctx iris.Context) {
	var req types.ShopOfferSummary
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}

	offer := orm.ShopOffer{
		ID:             req.ID,
		Effects:        req.Effects,
		Number:         req.Number,
		ResourceNumber: req.ResourceNumber,
		ResourceID:     req.ResourceID,
		Type:           req.Type,
	}

	if err := orm.GormDB.Create(&offer).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create shop offer", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

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

	offer.Effects = req.Effects
	offer.Number = req.Number
	offer.ResourceNumber = req.ResourceNumber
	offer.ResourceID = req.ResourceID
	offer.Type = req.Type

	if err := orm.GormDB.Save(&offer).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update shop offer", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

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
