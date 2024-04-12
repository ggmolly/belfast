package dto

import (
	"github.com/ggmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoBuildId struct {
	Id int `params:"build_id" validate:"required"`
}

// Places the DtoBuildId in dtoBuildId into the fiber context
func BuildIdMiddleware(ctx *fiber.Ctx) error {
	var buildId DtoBuildId

	if err := ctx.ParamsParser(&buildId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(buildId); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoBuildId", buildId)
	return ctx.Next()
}

type DtoBuildEdit struct {
	Id         uint32 `params:"build_id"`
	ShipId     uint32 `form:"template_id"`
	FinishesAt string `form:"finishes_at" validate:"datetime=2006-01-02T15:04:05"`
	Action     string `form:"action" validate:"required,oneof=save delete new duplicate finish"`
}

// Places the DtoBuildEdit in dtoBuildEdit into the fiber context
func BuildEditMiddleware(ctx *fiber.Ctx) error {
	var buildEdit DtoBuildEdit

	if err := ctx.BodyParser(&buildEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(buildEdit); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoBuildEdit", buildEdit)
	return ctx.Next()
}
