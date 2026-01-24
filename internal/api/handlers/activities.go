package handlers

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
)

const (
	activityAllowlistCategory = "ServerCfg/activities.json"
	activityAllowlistKey      = "allowlist"
)

type ActivityHandler struct{}

func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{}
}

func RegisterActivityRoutes(party iris.Party, handler *ActivityHandler) {
	party.Get("/allowlist", handler.GetAllowlist)
	party.Put("/allowlist", handler.ReplaceAllowlist)
	party.Patch("/allowlist", handler.UpdateAllowlist)
}

// GetAllowlist godoc
// @Summary     Get activity allowlist
// @Tags        Activities
// @Produce     json
// @Success     200  {object}  ActivityAllowlistResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/activities/allowlist [get]
func (handler *ActivityHandler) GetAllowlist(ctx iris.Context) {
	allowlist, err := loadActivityAllowlist()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load allowlist", nil))
		return
	}
	payload := types.ActivityAllowlistPayload{IDs: allowlist}
	_ = ctx.JSON(response.Success(payload))
}

// ReplaceAllowlist godoc
// @Summary     Replace activity allowlist
// @Tags        Activities
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ActivityAllowlistPayload  true  "Allowlist"
// @Success     200  {object}  ActivityAllowlistResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/activities/allowlist [put]
func (handler *ActivityHandler) ReplaceAllowlist(ctx iris.Context) {
	var req types.ActivityAllowlistPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	allowlist, err := validateActivityIDs(req.IDs)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := saveActivityAllowlist(allowlist); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to save allowlist", nil))
		return
	}
	payload := types.ActivityAllowlistPayload{IDs: allowlist}
	_ = ctx.JSON(response.Success(payload))
}

// UpdateAllowlist godoc
// @Summary     Patch activity allowlist
// @Tags        Activities
// @Accept      json
// @Produce     json
// @Param       payload  body      types.ActivityAllowlistPatchPayload  true  "Allowlist update"
// @Success     200  {object}  ActivityAllowlistResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/activities/allowlist [patch]
func (handler *ActivityHandler) UpdateAllowlist(ctx iris.Context) {
	var req types.ActivityAllowlistPatchPayload
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid request", nil))
		return
	}
	allowlist, err := loadActivityAllowlist()
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load allowlist", nil))
		return
	}
	set := make(map[uint32]struct{}, len(allowlist))
	for _, id := range allowlist {
		set[id] = struct{}{}
	}
	for _, id := range req.Add {
		set[id] = struct{}{}
	}
	for _, id := range req.Remove {
		delete(set, id)
	}
	merged := make([]uint32, 0, len(set))
	for id := range set {
		merged = append(merged, id)
	}
	validated, err := validateActivityIDs(merged)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	if err := saveActivityAllowlist(validated); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to save allowlist", nil))
		return
	}
	payload := types.ActivityAllowlistPayload{IDs: validated}
	_ = ctx.JSON(response.Success(payload))
}

func loadActivityAllowlist() ([]uint32, error) {
	var entry orm.ConfigEntry
	result := orm.GormDB.Where("category = ? AND key = ?", activityAllowlistCategory, activityAllowlistKey).Limit(1).Find(&entry)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return []uint32{}, nil
	}
	var allowlist []uint32
	if err := json.Unmarshal(entry.Data, &allowlist); err != nil {
		return nil, err
	}
	return allowlist, nil
}

func saveActivityAllowlist(ids []uint32) error {
	payload, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	entry := orm.ConfigEntry{Category: activityAllowlistCategory, Key: activityAllowlistKey, Data: payload}
	var existing orm.ConfigEntry
	result := orm.GormDB.Where("category = ? AND key = ?", activityAllowlistCategory, activityAllowlistKey).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return orm.GormDB.Create(&entry).Error
	}
	return orm.GormDB.Model(&existing).Update("data", payload).Error
}

func validateActivityIDs(ids []uint32) ([]uint32, error) {
	unique := make(map[uint32]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	validated := make([]uint32, 0, len(unique))
	for id := range unique {
		var entry orm.ConfigEntry
		result := orm.GormDB.Where("category = ? AND key = ?", "ShareCfg/activity_template.json", fmt.Sprintf("%d", id)).Limit(1).Find(&entry)
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected == 0 {
			return nil, fmt.Errorf("unknown activity id %d", id)
		}
		validated = append(validated, id)
	}
	sort.Slice(validated, func(i, j int) bool { return validated[i] < validated[j] })
	return validated, nil
}

func writeActivityError(ctx iris.Context, err error, resource string) {
	if err == nil {
		return
	}
	if err == gorm.ErrRecordNotFound {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", fmt.Sprintf("%s not found", resource), nil))
		return
	}
	ctx.StatusCode(iris.StatusInternalServerError)
	_ = ctx.JSON(response.Error("internal_error", fmt.Sprintf("failed to load %s", resource), nil))
}
