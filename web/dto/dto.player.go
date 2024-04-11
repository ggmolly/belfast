package dto

import (
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoPlayerId struct {
	Id int `params:"player_id" validate:"required"`
}

type DtoQuickPlayerEdit struct {
	Name  string `form:"name" validate:"required,min=3,max=30"`
	Level int    `form:"level" validate:"required,min=1,max=9999"`
}

// Places the DtoPlayerId in dtoPlayerId into the fiber context
func PlayerIdMiddleware(ctx *fiber.Ctx) error {
	var playerId DtoPlayerId

	if err := ctx.ParamsParser(&playerId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(playerId); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoPlayerId", playerId)
	return ctx.Next()
}

// Places the DtoQuickPlayerEdit in dtoQuickPlayerEdit into the fiber context
func QuickPlayerEditMiddleware(ctx *fiber.Ctx) error {
	var playerEdit DtoQuickPlayerEdit

	if err := ctx.BodyParser(&playerEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(playerEdit); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralAlert(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoQuickPlayerEdit", playerEdit)
	return ctx.Next()
}
