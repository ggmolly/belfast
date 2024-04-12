package dto

import (
	"github.com/ggmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoResourceId struct {
	ResourceId int `params:"resource_id" validate:"required"`
}

// Places the DtoResourceId in dtoResourceId into the fiber context
func ResourcesIdMiddleware(ctx *fiber.Ctx) error {
	var resourceId DtoResourceId

	if err := ctx.ParamsParser(&resourceId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(resourceId); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoResourceId", resourceId)
	return ctx.Next()
}

type DtoResourceEdit struct {
	Amount     uint32 `form:"amount" validate:"omitempty,min=1,max=4294967295"`
	Action     string `form:"action" validate:"required,oneof=save delete new"`
	ResourceId uint32 `form:"resource_id" validate:"omitempty"`
}

// Places the DtoResourceEdit in dtoResourceEdit into the fiber context
func ResourcesEditMiddleware(ctx *fiber.Ctx) error {
	var resourceEdit DtoResourceEdit

	if err := ctx.BodyParser(&resourceEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(resourceEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}
	ctx.Locals("dtoResourceEdit", resourceEdit)
	return ctx.Next()
}
