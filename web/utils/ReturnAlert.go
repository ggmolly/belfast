package utils

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	ALERT_COLOR_DEFAULT = "primary"
	ALERT_COLOR_SUCCESS = "success"
	ALERT_COLOR_WARNING = "warning"
	ALERT_COLOR_ERROR   = "error"
)

var (
	errInvalidColor = errors.New("invalid color")
)

func isColorValid(color string) bool {
	return color == ALERT_COLOR_DEFAULT || color == ALERT_COLOR_WARNING || color == ALERT_COLOR_ERROR || color == ALERT_COLOR_SUCCESS
}

func RenderEphemeralAlert(color, message string, ctx *fiber.Ctx) error {
	if !isColorValid(color) {
		return errInvalidColor
	}
	return ctx.Render(fmt.Sprintf("components/alerts/ephemeral/ephemeral_alert_%s", color), fiber.Map{
		"message": message,
	})
}

func RenderAlert(color, message string, ctx *fiber.Ctx) error {
	if !isColorValid(color) {
		return errInvalidColor
	}
	return ctx.Render(fmt.Sprintf("components/alerts/default/alert_%s", color), fiber.Map{
		"message": message,
	})
}
