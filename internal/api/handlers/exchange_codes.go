package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

type ExchangeCodeHandler struct{}

func NewExchangeCodeHandler() *ExchangeCodeHandler {
	return &ExchangeCodeHandler{}
}

func RegisterExchangeCodeRoutes(party iris.Party, handler *ExchangeCodeHandler) {
	party.Get("", handler.ListExchangeCodes)
	party.Get("/{id:uint}", handler.ExchangeCodeDetail)
	party.Post("", handler.CreateExchangeCode)
	party.Put("/{id:uint}", handler.UpdateExchangeCode)
	party.Delete("/{id:uint}", handler.DeleteExchangeCode)
}

// ListExchangeCodes godoc
// @Summary     List exchange codes
// @Tags        Exchange Codes
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  ExchangeCodeListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/exchange-codes [get]
func (handler *ExchangeCodeHandler) ListExchangeCodes(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var total int64
	if err := orm.GormDB.Model(&orm.ExchangeCode{}).Count(&total).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list exchange codes", nil))
		return
	}

	var codes []orm.ExchangeCode
	if err := orm.GormDB.Order("id asc").Limit(pagination.Limit).Offset(pagination.Offset).Find(&codes).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list exchange codes", nil))
		return
	}

	results := make([]types.ExchangeCodeSummary, 0, len(codes))
	for _, code := range codes {
		rewards, err := decodeExchangeRewards(code.Rewards)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to decode rewards", nil))
			return
		}
		results = append(results, types.ExchangeCodeSummary{
			ID:       code.ID,
			Code:     code.Code,
			Platform: code.Platform,
			Quota:    code.Quota,
			Rewards:  rewards,
		})
	}

	payload := types.ExchangeCodeListResponse{
		Codes: results,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}

	_ = ctx.JSON(response.Success(payload))
}

// ExchangeCodeDetail godoc
// @Summary     Get exchange code
// @Tags        Exchange Codes
// @Produce     json
// @Param       id   path  int  true  "Exchange code ID"
// @Success     200  {object}  ExchangeCodeSummaryResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/exchange-codes/{id} [get]
func (handler *ExchangeCodeHandler) ExchangeCodeDetail(ctx iris.Context) {
	codeID, err := parsePathUint32(ctx.Params().Get("id"), "exchange code id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var code orm.ExchangeCode
	if err := orm.GormDB.First(&code, codeID).Error; err != nil {
		writeExchangeCodeError(ctx, err)
		return
	}

	rewards, err := decodeExchangeRewards(code.Rewards)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to decode rewards", nil))
		return
	}

	payload := types.ExchangeCodeSummary{
		ID:       code.ID,
		Code:     code.Code,
		Platform: code.Platform,
		Quota:    code.Quota,
		Rewards:  rewards,
	}

	_ = ctx.JSON(response.Success(payload))
}

// CreateExchangeCode godoc
// @Summary     Create exchange code
// @Tags        Exchange Codes
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ExchangeCodeRequest  true  "Exchange code"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/exchange-codes [post]
func (handler *ExchangeCodeHandler) CreateExchangeCode(ctx iris.Context) {
	req, err := readExchangeCodeRequest(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	code := orm.ExchangeCode{
		Code:     req.Code,
		Platform: req.Platform,
		Quota:    req.Quota,
		Rewards:  req.Rewards,
	}

	if err := orm.GormDB.Create(&code).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create exchange code", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// UpdateExchangeCode godoc
// @Summary     Update exchange code
// @Tags        Exchange Codes
// @Accept      json
// @Produce     json
// @Param       id       path      int                       true  "Exchange code ID"
// @Param       payload  body      types.ExchangeCodeRequest  true  "Exchange code"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/exchange-codes/{id} [put]
func (handler *ExchangeCodeHandler) UpdateExchangeCode(ctx iris.Context) {
	codeID, err := parsePathUint32(ctx.Params().Get("id"), "exchange code id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	req, err := readExchangeCodeRequest(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var code orm.ExchangeCode
	if err := orm.GormDB.First(&code, codeID).Error; err != nil {
		writeExchangeCodeError(ctx, err)
		return
	}

	code.Code = req.Code
	code.Platform = req.Platform
	code.Quota = req.Quota
	code.Rewards = req.Rewards

	if err := orm.GormDB.Save(&code).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update exchange code", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

// DeleteExchangeCode godoc
// @Summary     Delete exchange code
// @Tags        Exchange Codes
// @Produce     json
// @Param       id   path  int  true  "Exchange code ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/exchange-codes/{id} [delete]
func (handler *ExchangeCodeHandler) DeleteExchangeCode(ctx iris.Context) {
	codeID, err := parsePathUint32(ctx.Params().Get("id"), "exchange code id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}

	var code orm.ExchangeCode
	if err := orm.GormDB.First(&code, codeID).Error; err != nil {
		writeExchangeCodeError(ctx, err)
		return
	}

	transaction := orm.GormDB.Begin()
	if err := transaction.Where("exchange_code_id = ?", codeID).Delete(&orm.ExchangeCodeRedeem{}).Error; err != nil {
		transaction.Rollback()
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete exchange code redeems", nil))
		return
	}
	if err := transaction.Delete(&orm.ExchangeCode{}, codeID).Error; err != nil {
		transaction.Rollback()
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete exchange code", nil))
		return
	}
	if err := transaction.Commit().Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete exchange code", nil))
		return
	}

	_ = ctx.JSON(response.Success(nil))
}

func readExchangeCodeRequest(ctx iris.Context) (*exchangeCodeRequestPayload, error) {
	var req types.ExchangeCodeRequest
	if err := ctx.ReadJSON(&req); err != nil {
		return nil, fmt.Errorf("invalid request")
	}
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}
	quota := -1
	if req.Quota != nil {
		quota = *req.Quota
	}
	if quota < -1 {
		return nil, fmt.Errorf("quota must be >= -1")
	}
	rewards, err := json.Marshal(req.Rewards)
	if err != nil {
		return nil, fmt.Errorf("invalid rewards")
	}
	return &exchangeCodeRequestPayload{
		Code:     code,
		Platform: strings.TrimSpace(req.Platform),
		Quota:    quota,
		Rewards:  rewards,
	}, nil
}

type exchangeCodeRequestPayload struct {
	Code     string
	Platform string
	Quota    int
	Rewards  json.RawMessage
}

func decodeExchangeRewards(payload []byte) ([]types.ExchangeReward, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	var rewards []types.ExchangeReward
	if err := json.Unmarshal(payload, &rewards); err != nil {
		return nil, err
	}
	return rewards, nil
}

func writeExchangeCodeError(ctx iris.Context, err error) {
	if err == gorm.ErrRecordNotFound {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "exchange code not found", nil))
		return
	}
	ctx.StatusCode(iris.StatusInternalServerError)
	_ = ctx.JSON(response.Error("internal_error", "failed to load exchange code", nil))
}
