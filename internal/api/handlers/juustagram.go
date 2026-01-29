package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"

	"github.com/ggmolly/belfast/internal/answer"
	"github.com/ggmolly/belfast/internal/api/response"
	"github.com/ggmolly/belfast/internal/api/types"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
)

type JuustagramHandler struct{}

func NewJuustagramHandler() *JuustagramHandler {
	return &JuustagramHandler{}
}

func RegisterJuustagramRoutes(party iris.Party, handler *JuustagramHandler) {
	party.Get("/templates", handler.ListTemplates)
	party.Post("/templates", handler.CreateTemplate)
	party.Put("/templates", handler.UpdateTemplate)
	party.Delete("/templates", handler.DeleteTemplate)
	party.Get("/templates/{id:uint}", handler.TemplateDetail)
	party.Get("/npc-templates", handler.ListNpcTemplates)
	party.Post("/npc-templates", handler.CreateNpcTemplate)
	party.Put("/npc-templates", handler.UpdateNpcTemplate)
	party.Delete("/npc-templates", handler.DeleteNpcTemplate)
	party.Get("/npc-templates/{id:uint}", handler.NpcTemplateDetail)
	party.Get("/ship-groups", handler.ListShipGroups)
	party.Post("/ship-groups", handler.CreateShipGroup)
	party.Put("/ship-groups", handler.UpdateShipGroup)
	party.Delete("/ship-groups", handler.DeleteShipGroup)
	party.Get("/ship-groups/{id:uint}", handler.ShipGroupDetail)
	party.Get("/language", handler.ListLanguage)
	party.Post("/language", handler.CreateLanguage)
	party.Put("/language", handler.UpdateLanguage)
	party.Delete("/language", handler.DeleteLanguage)
	party.Get("/language/{key:string}", handler.LanguageDetail)
}

func RegisterJuustagramPlayerRoutes(party iris.Party, handler *JuustagramHandler) {
	party.Get("/messages", handler.ListMessages)
	party.Get("/messages/{message_id:uint}", handler.MessageDetail)
	party.Patch("/messages/{message_id:uint}", handler.UpdateMessage)
	party.Get("/messages/{message_id:uint}/discuss", handler.MessageDiscussDetail)
	party.Post("/messages/{message_id:uint}/discuss", handler.MessageDiscussReply)
	party.Get("/message-states", handler.ListMessageStates)
	party.Get("/message-states/{message_id:uint}", handler.MessageStateDetail)
	party.Put("/message-states/{message_id:uint}", handler.UpdateMessageState)
	party.Delete("/message-states/{message_id:uint}", handler.DeleteMessageState)
	party.Get("/discuss/{message_id:uint}", handler.ListMessageDiscussStates)
	party.Put("/discuss/{message_id:uint}/{discuss_id:uint}", handler.UpsertMessageDiscussState)
	party.Delete("/discuss/{message_id:uint}/{discuss_id:uint}", handler.DeleteMessageDiscussState)
	party.Get("/groups", handler.ListChatGroups)
	party.Get("/groups/{group_id:uint}", handler.ChatGroupDetail)
	party.Post("/groups", handler.CreateChatGroup)
	party.Patch("/groups/{group_id:uint}", handler.UpdateChatGroup)
	party.Delete("/groups/{group_id:uint}", handler.DeleteChatGroup)
	party.Post("/groups/{group_id:uint}/chat-groups", handler.CreateChatGroupTopic)
	party.Delete("/chat-groups/{chat_group_id:uint}", handler.DeleteChatGroupTopic)
	party.Post("/chat-groups/{chat_group_id:uint}/reply", handler.CreateChatReply)
	party.Delete("/chat-groups/{chat_group_id:uint}/replies/{sequence:uint}", handler.DeleteChatReply)
	party.Patch("/chat-groups/read", handler.MarkChatGroupsRead)
}

// ListTemplates godoc
// @Summary     List Juustagram templates
// @Tags        Juustagram
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  JuustagramTemplateListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/templates [get]
func (handler *JuustagramHandler) ListTemplates(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	templates, total, err := orm.ListJuustagramTemplates(pagination.Offset, pagination.Limit)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list juustagram templates", nil))
		return
	}
	payload := types.JuustagramTemplateListResponse{
		Templates: templates,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}
	_ = ctx.JSON(response.Success(payload))
}

// TemplateDetail godoc
// @Summary     Get Juustagram template
// @Tags        Juustagram
// @Produce     json
// @Param       id   path      int  true  "Template ID"
// @Success     200  {object}  JuustagramTemplateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/templates/{id} [get]
func (handler *JuustagramHandler) TemplateDetail(ctx iris.Context) {
	messageID, err := parsePathUint32(ctx.Params().Get("id"), "template id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var template orm.JuustagramTemplate
	if err := orm.GormDB.First(&template, "id = ?", messageID).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram template")
		return
	}
	_ = ctx.JSON(response.Success(template))
}

// CreateTemplate godoc
// @Summary     Create Juustagram template
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramTemplate  true  "Juustagram template"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/templates [post]
func (handler *JuustagramHandler) CreateTemplate(ctx iris.Context) {
	var req types.JuustagramTemplate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	name := strings.TrimSpace(req.Name)
	sculpture := strings.TrimSpace(req.Sculpture)
	picturePersist := strings.TrimSpace(req.PicturePersist)
	messagePersist := strings.TrimSpace(req.MessagePersist)
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
	if sculpture == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "sculpture is required", nil))
		return
	}
	if picturePersist == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "picture_persist is required", nil))
		return
	}
	if messagePersist == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "message_persist is required", nil))
		return
	}
	req.Name = name
	req.Sculpture = sculpture
	req.PicturePersist = picturePersist
	req.MessagePersist = messagePersist
	if err := orm.GormDB.Create(&req).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create juustagram template", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdateTemplate godoc
// @Summary     Update Juustagram template
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramTemplate  true  "Juustagram template"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/templates [put]
func (handler *JuustagramHandler) UpdateTemplate(ctx iris.Context) {
	var req types.JuustagramTemplate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	name := strings.TrimSpace(req.Name)
	sculpture := strings.TrimSpace(req.Sculpture)
	picturePersist := strings.TrimSpace(req.PicturePersist)
	messagePersist := strings.TrimSpace(req.MessagePersist)
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
	if sculpture == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "sculpture is required", nil))
		return
	}
	if picturePersist == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "picture_persist is required", nil))
		return
	}
	if messagePersist == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "message_persist is required", nil))
		return
	}
	var template orm.JuustagramTemplate
	if err := orm.GormDB.First(&template, "id = ?", req.ID).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram template")
		return
	}
	template.GroupID = req.GroupID
	template.ShipGroup = req.ShipGroup
	template.Name = name
	template.Sculpture = sculpture
	template.PicturePersist = picturePersist
	template.MessagePersist = messagePersist
	template.IsActive = req.IsActive
	template.NpcDiscussPersist = req.NpcDiscussPersist
	template.Time = req.Time
	template.TimePersist = req.TimePersist
	if err := orm.GormDB.Save(&template).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update juustagram template", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeleteTemplate godoc
// @Summary     Delete Juustagram template
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramTemplateDeleteRequest  true  "Template ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/templates [delete]
func (handler *JuustagramHandler) DeleteTemplate(ctx iris.Context) {
	var req types.JuustagramTemplateDeleteRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramTemplate{}, "id = ?", req.ID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete juustagram template", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram template not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// ListNpcTemplates godoc
// @Summary     List Juustagram NPC templates
// @Tags        Juustagram
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  JuustagramNpcTemplateListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/npc-templates [get]
func (handler *JuustagramHandler) ListNpcTemplates(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	templates, total, err := orm.ListJuustagramNpcTemplates(pagination.Offset, pagination.Limit)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list juustagram npc templates", nil))
		return
	}
	payload := types.JuustagramNpcTemplateListResponse{
		Templates: templates,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}
	_ = ctx.JSON(response.Success(payload))
}

// NpcTemplateDetail godoc
// @Summary     Get Juustagram NPC template
// @Tags        Juustagram
// @Produce     json
// @Param       id   path      int  true  "NPC Template ID"
// @Success     200  {object}  JuustagramNpcTemplateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/npc-templates/{id} [get]
func (handler *JuustagramHandler) NpcTemplateDetail(ctx iris.Context) {
	templateID, err := parsePathUint32(ctx.Params().Get("id"), "npc template id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var template orm.JuustagramNpcTemplate
	if err := orm.GormDB.First(&template, "id = ?", templateID).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram npc template")
		return
	}
	_ = ctx.JSON(response.Success(template))
}

// CreateNpcTemplate godoc
// @Summary     Create Juustagram NPC template
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramNpcTemplate  true  "Juustagram NPC template"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/npc-templates [post]
func (handler *JuustagramHandler) CreateNpcTemplate(ctx iris.Context) {
	var req types.JuustagramNpcTemplate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	messagePersist := strings.TrimSpace(req.MessagePersist)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if messagePersist == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "message_persist is required", nil))
		return
	}
	req.MessagePersist = messagePersist
	if err := orm.GormDB.Create(&req).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create juustagram npc template", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdateNpcTemplate godoc
// @Summary     Update Juustagram NPC template
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramNpcTemplate  true  "Juustagram NPC template"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/npc-templates [put]
func (handler *JuustagramHandler) UpdateNpcTemplate(ctx iris.Context) {
	var req types.JuustagramNpcTemplate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	messagePersist := strings.TrimSpace(req.MessagePersist)
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	if messagePersist == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "message_persist is required", nil))
		return
	}
	var template orm.JuustagramNpcTemplate
	if err := orm.GormDB.First(&template, "id = ?", req.ID).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram npc template")
		return
	}
	template.ShipGroup = req.ShipGroup
	template.MessagePersist = messagePersist
	template.NpcReplyPersist = req.NpcReplyPersist
	template.TimePersist = req.TimePersist
	if err := orm.GormDB.Save(&template).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update juustagram npc template", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeleteNpcTemplate godoc
// @Summary     Delete Juustagram NPC template
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramNpcTemplateDeleteRequest  true  "NPC Template ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/npc-templates [delete]
func (handler *JuustagramHandler) DeleteNpcTemplate(ctx iris.Context) {
	var req types.JuustagramNpcTemplateDeleteRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if req.ID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "id is required", nil))
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramNpcTemplate{}, "id = ?", req.ID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete juustagram npc template", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram npc template not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// ListShipGroups godoc
// @Summary     List Juustagram ship groups
// @Tags        Juustagram
// @Produce     json
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  JuustagramShipGroupListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/ship-groups [get]
func (handler *JuustagramHandler) ListShipGroups(ctx iris.Context) {
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	groups, total, err := orm.ListJuustagramShipGroupTemplates(pagination.Offset, pagination.Limit)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list juustagram ship groups", nil))
		return
	}
	payload := types.JuustagramShipGroupListResponse{
		Groups: groups,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}
	_ = ctx.JSON(response.Success(payload))
}

// ShipGroupDetail godoc
// @Summary     Get Juustagram ship group
// @Tags        Juustagram
// @Produce     json
// @Param       id   path      int  true  "Ship Group ID"
// @Success     200  {object}  JuustagramShipGroupResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/ship-groups/{id} [get]
func (handler *JuustagramHandler) ShipGroupDetail(ctx iris.Context) {
	shipGroup, err := parsePathUint32(ctx.Params().Get("id"), "ship group id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var template orm.JuustagramShipGroupTemplate
	if err := orm.GormDB.First(&template, "ship_group = ?", shipGroup).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram ship group")
		return
	}
	_ = ctx.JSON(response.Success(template))
}

// CreateShipGroup godoc
// @Summary     Create Juustagram ship group
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramShipGroupTemplate  true  "Juustagram ship group"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/ship-groups [post]
func (handler *JuustagramHandler) CreateShipGroup(ctx iris.Context) {
	var req types.JuustagramShipGroupTemplate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	name := strings.TrimSpace(req.Name)
	background := strings.TrimSpace(req.Background)
	sculpture := strings.TrimSpace(req.Sculpture)
	sculptureII := strings.TrimSpace(req.SculptureII)
	if req.ShipGroup == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "ship_group is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}
	if background == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "background is required", nil))
		return
	}
	if sculpture == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "sculpture is required", nil))
		return
	}
	if sculptureII == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "sculpture_ii is required", nil))
		return
	}
	req.Name = name
	req.Background = background
	req.Sculpture = sculpture
	req.SculptureII = sculptureII
	if err := orm.GormDB.Create(&req).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create juustagram ship group", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdateShipGroup godoc
// @Summary     Update Juustagram ship group
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramShipGroupTemplate  true  "Juustagram ship group"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/ship-groups [put]
func (handler *JuustagramHandler) UpdateShipGroup(ctx iris.Context) {
	var req types.JuustagramShipGroupTemplate
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	name := strings.TrimSpace(req.Name)
	background := strings.TrimSpace(req.Background)
	sculpture := strings.TrimSpace(req.Sculpture)
	sculptureII := strings.TrimSpace(req.SculptureII)
	if req.ShipGroup == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "ship_group is required", nil))
		return
	}
	if name == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "name is required", nil))
		return
	}
	if background == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "background is required", nil))
		return
	}
	if sculpture == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "sculpture is required", nil))
		return
	}
	if sculptureII == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "sculpture_ii is required", nil))
		return
	}
	var template orm.JuustagramShipGroupTemplate
	if err := orm.GormDB.First(&template, "ship_group = ?", req.ShipGroup).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram ship group")
		return
	}
	template.Name = name
	template.Background = background
	template.Sculpture = sculpture
	template.SculptureII = sculptureII
	template.Nationality = req.Nationality
	template.Type = req.Type
	if err := orm.GormDB.Save(&template).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update juustagram ship group", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeleteShipGroup godoc
// @Summary     Delete Juustagram ship group
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramShipGroupDeleteRequest  true  "Ship group"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/ship-groups [delete]
func (handler *JuustagramHandler) DeleteShipGroup(ctx iris.Context) {
	var req types.JuustagramShipGroupDeleteRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if req.ShipGroup == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "ship_group is required", nil))
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramShipGroupTemplate{}, "ship_group = ?", req.ShipGroup)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete juustagram ship group", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram ship group not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// LanguageDetail godoc
// @Summary     Get Juustagram language entry
// @Tags        Juustagram
// @Produce     json
// @Param       key   path      string  true  "Language key"
// @Success     200  {object}  JuustagramLanguageResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/language/{key} [get]
func (handler *JuustagramHandler) LanguageDetail(ctx iris.Context) {
	key := ctx.Params().Get("key")
	if key == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "language key is required", nil))
		return
	}
	var entry orm.JuustagramLanguage
	if err := orm.GormDB.First(&entry, "key = ?", key).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram language")
		return
	}
	_ = ctx.JSON(response.Success(entry))
}

// ListLanguage godoc
// @Summary     List Juustagram language entries
// @Tags        Juustagram
// @Produce     json
// @Param       prefix  query  string  false  "Language key prefix"
// @Success     200  {object}  JuustagramLanguageListResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/language [get]
func (handler *JuustagramHandler) ListLanguage(ctx iris.Context) {
	prefix := strings.TrimSpace(ctx.URLParam("prefix"))
	entries, err := orm.ListJuustagramLanguageByPrefix(prefix)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list juustagram language entries", nil))
		return
	}
	payload := types.JuustagramLanguageListResponse{Entries: entries}
	_ = ctx.JSON(response.Success(payload))
}

// CreateLanguage godoc
// @Summary     Create Juustagram language entry
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramLanguage  true  "Language entry"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/language [post]
func (handler *JuustagramHandler) CreateLanguage(ctx iris.Context) {
	var req types.JuustagramLanguage
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	key := strings.TrimSpace(req.Key)
	value := strings.TrimSpace(req.Value)
	if key == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "key is required", nil))
		return
	}
	if value == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "value is required", nil))
		return
	}
	req.Key = key
	req.Value = value
	if err := orm.GormDB.Create(&req).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create juustagram language entry", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// UpdateLanguage godoc
// @Summary     Update Juustagram language entry
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramLanguage  true  "Language entry"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/language [put]
func (handler *JuustagramHandler) UpdateLanguage(ctx iris.Context) {
	var req types.JuustagramLanguage
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	key := strings.TrimSpace(req.Key)
	value := strings.TrimSpace(req.Value)
	if key == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "key is required", nil))
		return
	}
	if value == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "value is required", nil))
		return
	}
	var entry orm.JuustagramLanguage
	if err := orm.GormDB.First(&entry, "key = ?", key).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram language")
		return
	}
	entry.Value = value
	if err := orm.GormDB.Save(&entry).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update juustagram language entry", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// DeleteLanguage godoc
// @Summary     Delete Juustagram language entry
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       payload  body      types.JuustagramLanguageDeleteRequest  true  "Language key"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/juustagram/language [delete]
func (handler *JuustagramHandler) DeleteLanguage(ctx iris.Context) {
	var req types.JuustagramLanguageDeleteRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "key is required", nil))
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramLanguage{}, "key = ?", key)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete juustagram language entry", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram language not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// ListMessages godoc
// @Summary     List Juustagram messages
// @Tags        Juustagram
// @Produce     json
// @Param       id      path  int  true  "Commander ID"
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  JuustagramMessageListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/messages [get]
func (handler *JuustagramHandler) ListMessages(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	templates, total, err := orm.ListJuustagramTemplates(pagination.Offset, pagination.Limit)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list juustagram messages", nil))
		return
	}
	now := uint32(time.Now().Unix())
	messages := make([]types.JuustagramMessage, 0, len(templates))
	for _, template := range templates {
		payload, err := answer.BuildJuustagramMessage(commander.CommanderID, template.ID, now)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to build juustagram message", nil))
			return
		}
		messages = append(messages, juustagramMessageFromProto(payload))
	}
	responsePayload := types.JuustagramMessageListResponse{
		Messages: messages,
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}
	_ = ctx.JSON(response.Success(responsePayload))
}

// MessageDetail godoc
// @Summary     Get Juustagram message
// @Tags        Juustagram
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Success     200  {object}  JuustagramMessageResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/messages/{message_id} [get]
func (handler *JuustagramHandler) MessageDetail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	now := uint32(time.Now().Unix())
	message, err := answer.BuildJuustagramMessage(commander.CommanderID, messageID, now)
	if err != nil {
		writeGameDataError(ctx, err, "juustagram message")
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramMessageResponse{Message: juustagramMessageFromProto(message)}))
}

// UpdateMessage godoc
// @Summary     Update Juustagram message state
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Param       payload     body  types.JuustagramMessageUpdateRequest  true  "Message update"
// @Success     200  {object}  JuustagramMessageResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/messages/{message_id} [patch]
func (handler *JuustagramHandler) UpdateMessage(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.JuustagramMessageUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	now := uint32(time.Now().Unix())
	state, err := orm.GetOrCreateJuustagramMessageState(commander.CommanderID, messageID, now)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load message state", nil))
		return
	}
	state.UpdatedAt = now
	if req.Read != nil {
		if *req.Read {
			state.IsRead = 1
		} else {
			state.IsRead = 0
		}
	}
	if req.Like != nil {
		if *req.Like {
			if state.IsGood == 0 {
				state.IsGood = 1
				state.GoodCount += 1
			}
		} else {
			if state.IsGood == 1 {
				state.IsGood = 0
				if state.GoodCount > 0 {
					state.GoodCount -= 1
				}
			}
		}
	}
	if err := orm.SaveJuustagramMessageState(state); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update message state", nil))
		return
	}
	message, err := answer.BuildJuustagramMessage(commander.CommanderID, messageID, now)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build juustagram message", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramMessageResponse{Message: juustagramMessageFromProto(message)}))
}

// ListMessageStates godoc
// @Summary     List Juustagram message states
// @Tags        Juustagram
// @Produce     json
// @Param       id   path  int  true  "Commander ID"
// @Success     200  {object}  JuustagramMessageStateListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/message-states [get]
func (handler *JuustagramHandler) ListMessageStates(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var states []orm.JuustagramMessageState
	if err := orm.GormDB.Where("commander_id = ?", commander.CommanderID).Order("message_id asc").Find(&states).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load message states", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramMessageStateListResponse{States: states}))
}

// MessageStateDetail godoc
// @Summary     Get Juustagram message state
// @Tags        Juustagram
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Success     200  {object}  JuustagramMessageStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/message-states/{message_id} [get]
func (handler *JuustagramHandler) MessageStateDetail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	state, err := orm.GetJuustagramMessageState(commander.CommanderID, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "juustagram message state not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load message state", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramMessageStateResponse{State: *state}))
}

// UpdateMessageState godoc
// @Summary     Update Juustagram message state
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Param       payload     body  types.JuustagramMessageStateUpdateRequest  true  "Message state update"
// @Success     200  {object}  JuustagramMessageStateResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/message-states/{message_id} [put]
func (handler *JuustagramHandler) UpdateMessageState(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.JuustagramMessageStateUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	state, err := orm.GetJuustagramMessageState(commander.CommanderID, messageID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load message state", nil))
			return
		}
		state = &orm.JuustagramMessageState{CommanderID: commander.CommanderID, MessageID: messageID}
	}
	state.IsRead = req.IsRead
	state.IsGood = req.IsGood
	state.GoodCount = req.GoodCount
	state.UpdatedAt = req.UpdatedAt
	if err := orm.SaveJuustagramMessageState(state); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update message state", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramMessageStateResponse{State: *state}))
}

// DeleteMessageState godoc
// @Summary     Delete Juustagram message state
// @Tags        Juustagram
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/message-states/{message_id} [delete]
func (handler *JuustagramHandler) DeleteMessageState(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramMessageState{}, "commander_id = ? AND message_id = ?", commander.CommanderID, messageID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete message state", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram message state not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// ListMessageDiscussStates godoc
// @Summary     List Juustagram discuss selections
// @Tags        Juustagram
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Success     200  {object}  JuustagramPlayerDiscussListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/discuss/{message_id} [get]
func (handler *JuustagramHandler) ListMessageDiscussStates(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	entries, err := orm.ListJuustagramPlayerDiscuss(commander.CommanderID, messageID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load discuss selections", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramPlayerDiscussListResponse{Entries: entries}))
}

// UpsertMessageDiscussState godoc
// @Summary     Update Juustagram discuss selection
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Param       discuss_id  path  int  true  "Discuss ID"
// @Param       payload     body  types.JuustagramPlayerDiscussUpdateRequest  true  "Discuss selection"
// @Success     200  {object}  JuustagramPlayerDiscussResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/discuss/{message_id}/{discuss_id} [put]
func (handler *JuustagramHandler) UpsertMessageDiscussState(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	discussID, err := parsePathUint32(ctx.Params().Get("discuss_id"), "discuss id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.JuustagramPlayerDiscussUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	entry, err := orm.GetJuustagramPlayerDiscuss(commander.CommanderID, messageID, discussID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load discuss selection", nil))
			return
		}
		entry = &orm.JuustagramPlayerDiscuss{CommanderID: commander.CommanderID, MessageID: messageID, DiscussID: discussID}
	}
	entry.OptionIndex = req.OptionIndex
	entry.NpcReplyID = req.NpcReplyID
	entry.CommentTime = req.CommentTime
	if err := orm.UpsertJuustagramPlayerDiscuss(entry); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to save discuss selection", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramPlayerDiscussResponse{Entry: *entry}))
}

// DeleteMessageDiscussState godoc
// @Summary     Delete Juustagram discuss selection
// @Tags        Juustagram
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Param       discuss_id  path  int  true  "Discuss ID"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/discuss/{message_id}/{discuss_id} [delete]
func (handler *JuustagramHandler) DeleteMessageDiscussState(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	discussID, err := parsePathUint32(ctx.Params().Get("discuss_id"), "discuss id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramPlayerDiscuss{}, "commander_id = ? AND message_id = ? AND discuss_id = ?", commander.CommanderID, messageID, discussID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete discuss selection", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram discuss selection not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// MessageDiscussDetail godoc
// @Summary     Get Juustagram message discuss options
// @Tags        Juustagram
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Success     200  {object}  JuustagramDiscussResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/messages/{message_id}/discuss [get]
func (handler *JuustagramHandler) MessageDiscussDetail(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	options, err := answer.ListJuustagramDiscussOptions(messageID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load discuss options", nil))
		return
	}
	selection, err := orm.ListJuustagramPlayerDiscuss(commander.CommanderID, messageID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load discuss selection", nil))
		return
	}
	selectionsPayload := make([]types.JuustagramDiscussSelection, 0, len(selection))
	for _, selected := range selection {
		selectionsPayload = append(selectionsPayload, types.JuustagramDiscussSelection{
			DiscussID:   selected.DiscussID,
			OptionIndex: selected.OptionIndex,
			NpcReplyID:  selected.NpcReplyID,
			CommentTime: selected.CommentTime,
		})
	}
	optionsPayload := make([]types.JuustagramDiscussOption, 0, len(options))
	for _, option := range options {
		optionsPayload = append(optionsPayload, types.JuustagramDiscussOption{
			DiscussID:  option.DiscussID,
			Index:      option.Index,
			Text:       option.Text,
			NpcReplyID: option.NpcReplyID,
		})
	}
	_ = ctx.JSON(response.Success(types.JuustagramDiscussResponse{
		Options:    optionsPayload,
		Selections: selectionsPayload,
	}))
}

// MessageDiscussReply godoc
// @Summary     Reply to Juustagram message
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id          path  int  true  "Commander ID"
// @Param       message_id  path  int  true  "Message ID"
// @Param       payload     body  types.JuustagramDiscussRequest  true  "Discuss reply"
// @Success     200  {object}  JuustagramMessageResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/messages/{message_id}/discuss [post]
func (handler *JuustagramHandler) MessageDiscussReply(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	messageID, err := parsePathUint32(ctx.Params().Get("message_id"), "message id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.JuustagramDiscussRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	option, err := juustagramDiscussOption(messageID, req.DiscussID, req.OptionIndex)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	now := uint32(time.Now().Unix())
	entry, err := orm.GetJuustagramPlayerDiscuss(commander.CommanderID, messageID, req.DiscussID)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to load discuss selection", nil))
			return
		}
		entry = &orm.JuustagramPlayerDiscuss{}
	}
	entry.CommanderID = commander.CommanderID
	entry.MessageID = messageID
	entry.DiscussID = req.DiscussID
	entry.OptionIndex = req.OptionIndex
	entry.NpcReplyID = option.NpcReplyID
	entry.CommentTime = now
	if err := orm.UpsertJuustagramPlayerDiscuss(entry); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to save discuss reply", nil))
		return
	}
	message, err := answer.BuildJuustagramMessage(commander.CommanderID, messageID, now)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to build juustagram message", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramMessageResponse{Message: juustagramMessageFromProto(message)}))
}

// ListChatGroups godoc
// @Summary     List Juustagram chat groups
// @Tags        Juustagram
// @Produce     json
// @Param       id      path  int  true  "Commander ID"
// @Param       offset  query  int  false  "Pagination offset"
// @Param       limit   query  int  false  "Pagination limit"
// @Success     200  {object}  JuustagramGroupListResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/groups [get]
func (handler *JuustagramHandler) ListChatGroups(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	pagination, err := parsePagination(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	groups, total, err := orm.ListJuustagramGroups(commander.CommanderID, pagination.Offset, pagination.Limit)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to list juustagram groups", nil))
		return
	}
	payload := types.JuustagramGroupListResponse{
		Groups: juustagramGroupsFromModels(groups),
		Meta: types.PaginationMeta{
			Offset: pagination.Offset,
			Limit:  pagination.Limit,
			Total:  total,
		},
	}
	_ = ctx.JSON(response.Success(payload))
}

// ChatGroupDetail godoc
// @Summary     Get Juustagram chat group
// @Tags        Juustagram
// @Produce     json
// @Param       id        path  int  true  "Commander ID"
// @Param       group_id  path  int  true  "Group ID"
// @Success     200  {object}  JuustagramGroupResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/groups/{group_id} [get]
func (handler *JuustagramHandler) ChatGroupDetail(ctx iris.Context) {
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
	group, err := orm.GetJuustagramGroup(commander.CommanderID, groupID)
	if err != nil {
		writeGameDataError(ctx, err, "juustagram group")
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramGroupResponse{Group: juustagramGroupFromModel(*group)}))
}

// CreateChatGroup godoc
// @Summary     Create Juustagram group
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Commander ID"
// @Param       payload  body  types.JuustagramGroupCreateRequest  true  "Juustagram group"
// @Success     200  {object}  JuustagramGroupResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/groups [post]
func (handler *JuustagramHandler) CreateChatGroup(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.JuustagramGroupCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if req.GroupID == 0 || req.ChatGroupID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "group_id and chat_group_id are required", nil))
		return
	}
	group, err := orm.CreateJuustagramGroup(commander.CommanderID, req.GroupID, req.ChatGroupID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to create juustagram group", nil))
		return
	}
	if req.SkinID != 0 || req.Favorite != 0 {
		skinID := req.SkinID
		favorite := req.Favorite
		curGroup := req.ChatGroupID
		if err := orm.UpdateJuustagramGroup(commander.CommanderID, req.GroupID, &skinID, &favorite, &curGroup); err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to update juustagram group", nil))
			return
		}
		group, err = orm.GetJuustagramGroup(commander.CommanderID, req.GroupID)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(response.Error("internal_error", "failed to reload juustagram group", nil))
			return
		}
	}
	_ = ctx.JSON(response.Success(types.JuustagramGroupResponse{Group: juustagramGroupFromModel(*group)}))
}

// UpdateChatGroup godoc
// @Summary     Update Juustagram group settings
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Commander ID"
// @Param       group_id  path  int  true  "Group ID"
// @Param       payload   body  types.JuustagramGroupUpdateRequest  true  "Juustagram group update"
// @Success     200  {object}  JuustagramGroupResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/groups/{group_id} [patch]
func (handler *JuustagramHandler) UpdateChatGroup(ctx iris.Context) {
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
	var req types.JuustagramGroupUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if err := orm.UpdateJuustagramGroup(commander.CommanderID, groupID, req.SkinID, req.Favorite, req.CurChatGroup); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "juustagram group not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to update juustagram group", nil))
		return
	}
	group, err := orm.GetJuustagramGroup(commander.CommanderID, groupID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to reload juustagram group", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramGroupResponse{Group: juustagramGroupFromModel(*group)}))
}

// DeleteChatGroup godoc
// @Summary     Delete Juustagram group
// @Tags        Juustagram
// @Produce     json
// @Param       id        path  int  true  "Commander ID"
// @Param       group_id  path  int  true  "Group ID"
// @Success     200  {object}  JuustagramGroupDeleteResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/groups/{group_id} [delete]
func (handler *JuustagramHandler) DeleteChatGroup(ctx iris.Context) {
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
	result := orm.GormDB.Delete(&orm.JuustagramGroup{}, "commander_id = ? AND group_id = ?", commander.CommanderID, groupID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete juustagram group", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram group not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// CreateChatGroupTopic godoc
// @Summary     Activate Juustagram chat group
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id        path  int  true  "Commander ID"
// @Param       group_id  path  int  true  "Group ID"
// @Param       payload   body  types.JuustagramChatGroupCreateRequest  true  "Chat group"
// @Success     200  {object}  JuustagramGroupResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/groups/{group_id}/chat-groups [post]
func (handler *JuustagramHandler) CreateChatGroupTopic(ctx iris.Context) {
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
	var req types.JuustagramChatGroupCreateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if req.ChatGroupID == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "chat_group_id is required", nil))
		return
	}
	opTime := req.OpTime
	if opTime == 0 {
		opTime = orm.DefaultJuustagramOpTime()
	}
	if _, err := orm.CreateJuustagramChatGroup(commander.CommanderID, groupID, req.ChatGroupID, opTime); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "juustagram group not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to activate juustagram chat group", nil))
		return
	}
	group, err := orm.GetJuustagramGroup(commander.CommanderID, groupID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to reload juustagram group", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramGroupResponse{Group: juustagramGroupFromModel(*group)}))
}

// DeleteChatGroupTopic godoc
// @Summary     Delete Juustagram chat group
// @Tags        Juustagram
// @Produce     json
// @Param       id             path  int  true  "Commander ID"
// @Param       chat_group_id  path  int  true  "Chat Group ID"
// @Success     200  {object}  JuustagramChatGroupDeleteResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/chat-groups/{chat_group_id} [delete]
func (handler *JuustagramHandler) DeleteChatGroupTopic(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	chatGroupID, err := parsePathUint32(ctx.Params().Get("chat_group_id"), "chat group id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramChatGroup{}, "commander_id = ? AND chat_group_id = ?", commander.CommanderID, chatGroupID)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete juustagram chat group", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram chat group not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// CreateChatReply godoc
// @Summary     Reply to Juustagram chat group
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id             path  int  true  "Commander ID"
// @Param       chat_group_id  path  int  true  "Chat Group ID"
// @Param       payload        body  types.JuustagramChatReplyRequest  true  "Chat reply"
// @Success     200  {object}  JuustagramGroupResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/chat-groups/{chat_group_id}/reply [post]
func (handler *JuustagramHandler) CreateChatReply(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	chatGroupID, err := parsePathUint32(ctx.Params().Get("chat_group_id"), "chat group id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var req types.JuustagramChatReplyRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if req.ChatID == 0 || req.Value == 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "chat_id and value are required", nil))
		return
	}
	updated, err := orm.AddJuustagramChatReply(commander.CommanderID, chatGroupID, req.ChatID, req.Value, orm.DefaultJuustagramOpTime())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(response.Error("not_found", "juustagram chat group not found", nil))
			return
		}
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to add chat reply", nil))
		return
	}
	var group orm.JuustagramGroup
	if err := orm.GormDB.First(&group, "id = ? AND commander_id = ?", updated.GroupRecordID, commander.CommanderID).Error; err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to load juustagram group", nil))
		return
	}
	fullGroup, err := orm.GetJuustagramGroup(commander.CommanderID, group.GroupID)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to reload juustagram group", nil))
		return
	}
	_ = ctx.JSON(response.Success(types.JuustagramGroupResponse{Group: juustagramGroupFromModel(*fullGroup)}))
}

// DeleteChatReply godoc
// @Summary     Delete Juustagram chat reply
// @Tags        Juustagram
// @Produce     json
// @Param       id             path  int  true  "Commander ID"
// @Param       chat_group_id  path  int  true  "Chat Group ID"
// @Param       sequence       path  int  true  "Reply sequence"
// @Success     200  {object}  JuustagramChatReplyDeleteResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/chat-groups/{chat_group_id}/replies/{sequence} [delete]
func (handler *JuustagramHandler) DeleteChatReply(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	chatGroupID, err := parsePathUint32(ctx.Params().Get("chat_group_id"), "chat group id")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	sequence, err := parsePathUint32(ctx.Params().Get("sequence"), "reply sequence")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", err.Error(), nil))
		return
	}
	var chatGroup orm.JuustagramChatGroup
	if err := orm.GormDB.Select("id").Where("commander_id = ? AND chat_group_id = ?", commander.CommanderID, chatGroupID).First(&chatGroup).Error; err != nil {
		writeGameDataError(ctx, err, "juustagram chat group")
		return
	}
	result := orm.GormDB.Delete(&orm.JuustagramReply{}, "chat_group_record_id = ? AND sequence = ?", chatGroup.ID, sequence)
	if result.Error != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to delete chat reply", nil))
		return
	}
	if result.RowsAffected == 0 {
		ctx.StatusCode(iris.StatusNotFound)
		_ = ctx.JSON(response.Error("not_found", "juustagram chat reply not found", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

// MarkChatGroupsRead godoc
// @Summary     Mark Juustagram chat groups read
// @Tags        Juustagram
// @Accept      json
// @Produce     json
// @Param       id       path  int  true  "Commander ID"
// @Param       payload  body  types.JuustagramChatReadRequest  true  "Chat group IDs"
// @Success     200  {object}  OKResponseDoc
// @Failure     400  {object}  APIErrorResponseDoc
// @Failure     404  {object}  APIErrorResponseDoc
// @Failure     500  {object}  APIErrorResponseDoc
// @Router      /api/v1/players/{id}/juustagram/chat-groups/read [patch]
func (handler *JuustagramHandler) MarkChatGroupsRead(ctx iris.Context) {
	commander, err := loadCommanderDetail(ctx)
	if err != nil {
		writeCommanderError(ctx, err)
		return
	}
	var req types.JuustagramChatReadRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(response.Error("bad_request", "invalid payload", nil))
		return
	}
	if err := orm.SetJuustagramChatGroupRead(commander.CommanderID, req.ChatGroupIDs); err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(response.Error("internal_error", "failed to mark chat groups read", nil))
		return
	}
	_ = ctx.JSON(response.Success(nil))
}

func juustagramMessageFromProto(message *protobuf.INS_MESSAGE) types.JuustagramMessage {
	if message == nil {
		return types.JuustagramMessage{}
	}
	playerDiscuss := make([]types.JuustagramPlayerDiscuss, 0, len(message.GetPlayerDiscuss()))
	for _, entry := range message.GetPlayerDiscuss() {
		playerDiscuss = append(playerDiscuss, types.JuustagramPlayerDiscuss{
			ID:       entry.GetId(),
			Time:     entry.GetTime(),
			TextList: append([]string{}, entry.GetTextList()...),
			Text:     entry.GetText(),
			NpcReply: entry.GetNpcReply(),
		})
	}
	npcDiscuss := make([]types.JuustagramNpcComment, 0, len(message.GetNpcDiscuss()))
	for _, entry := range message.GetNpcDiscuss() {
		npcDiscuss = append(npcDiscuss, types.JuustagramNpcComment{
			ID:       entry.GetId(),
			Time:     entry.GetTime(),
			Text:     entry.GetText(),
			NpcReply: append([]uint32{}, entry.GetNpcReply()...),
		})
	}
	npcReply := make([]types.JuustagramNpcComment, 0, len(message.GetNpcReply()))
	for _, entry := range message.GetNpcReply() {
		npcReply = append(npcReply, types.JuustagramNpcComment{
			ID:       entry.GetId(),
			Time:     entry.GetTime(),
			Text:     entry.GetText(),
			NpcReply: append([]uint32{}, entry.GetNpcReply()...),
		})
	}
	return types.JuustagramMessage{
		ID:            message.GetId(),
		Time:          message.GetTime(),
		Text:          message.GetText(),
		Picture:       message.GetPicture(),
		PlayerDiscuss: playerDiscuss,
		NpcDiscuss:    npcDiscuss,
		NpcReply:      npcReply,
		Good:          message.GetGood(),
		IsGood:        message.GetIsGood(),
		IsRead:        message.GetIsRead(),
	}
}

func juustagramDiscussOption(messageID uint32, discussID uint32, index uint32) (answer.JuustagramDiscussOption, error) {
	options, err := answer.ListJuustagramDiscussOptions(messageID)
	if err != nil {
		return answer.JuustagramDiscussOption{}, err
	}
	for _, option := range options {
		if option.DiscussID == discussID && option.Index == index {
			return option, nil
		}
	}
	return answer.JuustagramDiscussOption{}, errors.New("invalid discuss option")
}

func juustagramGroupsFromModels(groups []orm.JuustagramGroup) []types.JuustagramGroup {
	response := make([]types.JuustagramGroup, 0, len(groups))
	for _, group := range groups {
		response = append(response, juustagramGroupFromModel(group))
	}
	return response
}

func juustagramGroupFromModel(group orm.JuustagramGroup) types.JuustagramGroup {
	chatGroups := make([]types.JuustagramChatGroup, 0, len(group.ChatGroups))
	for _, chatGroup := range group.ChatGroups {
		replies := make([]types.JuustagramReply, 0, len(chatGroup.ReplyList))
		for _, reply := range chatGroup.ReplyList {
			replies = append(replies, types.JuustagramReply{
				Sequence: reply.Sequence,
				Key:      reply.Key,
				Value:    reply.Value,
			})
		}
		chatGroups = append(chatGroups, types.JuustagramChatGroup{
			ChatGroupID: chatGroup.ChatGroupID,
			OpTime:      chatGroup.OpTime,
			ReadFlag:    chatGroup.ReadFlag,
			ReplyList:   replies,
		})
	}
	return types.JuustagramGroup{
		GroupID:      group.GroupID,
		SkinID:       group.SkinID,
		Favorite:     group.Favorite,
		CurChatGroup: group.CurChatGroup,
		ChatGroups:   chatGroups,
	}
}
