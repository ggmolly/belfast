package dto

import (
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoItemId struct {
	Id   int   `params:"item_id" validate:"required"`
	Data int64 `query:"data" validate:"omitempty"`
}

// Places the DtoItemId in dtoItemId into the fiber context
func ItemIdMiddleware(ctx *fiber.Ctx) error {
	var itemId DtoItemId

	if err := ctx.ParamsParser(&itemId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}
	if err := ctx.QueryParser(&itemId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(itemId); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoItemId", itemId)
	return ctx.Next()
}

type DtoItemEdit struct {
	Count      uint32 `form:"count" validate:"omitempty,min=1,max=4294967295"`
	Action     string `form:"action" validate:"required,oneof=save delete new"`
	TemplateId uint32 `form:"template_id" validate:"omitempty"`
	Data       uint32 `form:"data" validate:"omitempty"`
}

// Places the DtoItemEdit in dtoItemEdit into the fiber context
func ItemEditMiddleware(ctx *fiber.Ctx) error {
	var itemEdit DtoItemEdit

	if err := ctx.BodyParser(&itemEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(itemEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}
	ctx.Locals("dtoItemEdit", itemEdit)
	return ctx.Next()
}
