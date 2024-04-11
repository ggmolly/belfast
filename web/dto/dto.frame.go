package dto

import (
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoFrameId struct {
	Id int `query:"id" validate:"required"`
}

// Places the DtoFrameId in dtoFrameId into the fiber context
func FrameIdMiddleware(ctx *fiber.Ctx) error {
	var frameId DtoFrameId

	if err := ctx.QueryParser(&frameId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(frameId); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoFrameId", frameId)
	return ctx.Next()
}
