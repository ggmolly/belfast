package dto

import (
	"github.com/bettercallmolly/belfast/web/utils"
	"github.com/gofiber/fiber/v2"
)

type DtoAnnounceServer struct {
	Ip        string  `form:"server-ip" validate:"required,min=3,max=255"`
	Port      uint32  `form:"server-port" validate:"required,min=1,max=65535"`
	State     uint32  `form:"server-state" validate:"max=3"`
	ProxyIp   *string `form:"proxy-ip" omitempty validate:"omitempty,min=3,max=255"`
	ProxyPort *int    `form:"proxy-port" omitempty validate:"omitempty,min=1,max=65535"`
	Name      string  `form:"server-name" validate:"required,min=1,max=30"`
}

// Places the DtoServerData in dtoServerData into the fiber context
func ServerDataMiddleware(ctx *fiber.Ctx) error {
	var serverData DtoAnnounceServer

	if err := ctx.BodyParser(&serverData); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	if err := Validate.Struct(serverData); err != nil {
		ctx.Status(400)
		return utils.RenderEphemeralToast(utils.ALERT_COLOR_ERROR, err.Error(), ctx)
	}

	ctx.Locals("dtoServerData", serverData)
	return ctx.Next()
}
