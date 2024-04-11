package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func RenderEphemeralToast(color, message string, ctx *fiber.Ctx) error {
	if !isColorValid(color) {
		return errInvalidColor
	}
	return ctx.Render(fmt.Sprintf("components/toasts/ephemeral/ephemeral_toast_%s", color), fiber.Map{
		"message": message,
	})
}

func RenderToast(color, message string, ctx *fiber.Ctx) error {
	if !isColorValid(color) {
		return errInvalidColor
	}
	return ctx.Render(fmt.Sprintf("components/toasts/default/toast_%s", color), fiber.Map{
		"message": message,
	})
}
