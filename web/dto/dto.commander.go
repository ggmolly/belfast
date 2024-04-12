package dto

import (
	"github.com/ggmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoCommanderId struct {
	CommanderId uint32 `query:"commander_id" validate:"required,min=1"`
}

// Places the DtoCommanderId in dtoCommanderId into the fiber context
func CommanderIdMiddleware(ctx *fiber.Ctx) error {
	var commanderId DtoCommanderId

	if err := ctx.QueryParser(&commanderId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(commanderId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoCommanderId", commanderId)
	return ctx.Next()
}
