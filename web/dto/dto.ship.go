package dto

import (
	"time"

	"github.com/ggmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoShipId struct {
	Id int `params:"ship_id" validate:"required"`
}

// Places the DtoShipId in dtoShipId into the fiber context
func ShipIdMiddleware(ctx *fiber.Ctx) error {
	var shipId DtoShipId

	if err := ctx.ParamsParser(&shipId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(shipId); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoShipId", shipId)
	return ctx.Next()
}

type DtoShipEdit struct {
	Id             uint32    `params:"ship_id"`
	TemplateId     uint32    `form:"template_id"`
	Level          uint32    `form:"level" validate:"omitempty,min=1,max=125"`
	MaxLevel       uint32    `form:"max_level" validate:"omitempty,min=1,max=125"`
	Energy         uint32    `form:"energy" validate:"omitempty,min=0,max=150"`
	Intimacy       uint32    `form:"intimacy" validate:"omitempty,min=0,max=20000"`
	IsLocked       bool      `form:"locked"`
	IsSecretary    bool      `form:"secretary"`
	Propose        bool      `form:"propose"`
	CommonFlag     bool      `form:"common_flag"`
	BluePrintFlag  bool      `form:"blueprint_flag"`
	Proficiency    bool      `form:"proficiency"`
	ActivityNPC    uint32    `form:"activity_npc" validate:"omitempty,min=0"`
	CustomName     string    `form:"custom_name" validate:"omitempty,min=3,max=30"`
	CustomNameTime time.Time `form:"custom_name_time" validate:"omitempty"`
	CreateTime     string    `form:"create_time" validate:"datetime=2006-01-02T15:04:05"`
	SkinId         uint32    `form:"skin_id" validate:"omitempty,min=0"`
	Action         string    `form:"action" validate:"required,oneof=new save duplicate delete"`
}

// Places the DtoShipEdit in dtoShipEdit into the fiber context
func ShipEditMiddleware(ctx *fiber.Ctx) error {
	var shipEdit DtoShipEdit

	if err := ctx.BodyParser(&shipEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(shipEdit); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoShipEdit", shipEdit)
	return ctx.Next()
}

type DtoTemplateId struct {
	Id int `query:"template_id" validate:"required"`
}

// Places the DtoTemplateId in dtoTemplateId into the fiber context
func TemplateIdMiddleware(ctx *fiber.Ctx) error {
	var templateId DtoTemplateId

	if err := ctx.QueryParser(&templateId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(templateId); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoTemplateId", templateId)
	return ctx.Next()
}
