package dto

import (
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoServerId struct {
	Id int `params:"server_id" validate:"required"`
}

// Places the DtoServerId in dtoServerId into the fiber context
func ServerIdMiddleware(ctx *fiber.Ctx) error {
	var serverId DtoServerId

	if err := ctx.ParamsParser(&serverId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(serverId); err != nil {
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoServerId", serverId)
	return ctx.Next()
}
