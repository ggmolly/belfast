package middleware

import (
	"net/http"
	"strings"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/api/response"
)

func RegisterErrorHandlers(app *iris.Application) {
	app.OnAnyErrorCode(func(ctx iris.Context) {
		status := ctx.GetStatusCode()
		message := http.StatusText(status)
		if err := ctx.GetErr(); err != nil {
			message = err.Error()
		}
		code := errorCode(status)
		_ = ctx.JSON(response.Error(code, message, nil))
	})
}

func errorCode(status int) string {
	switch status {
	case iris.StatusBadRequest:
		return "bad_request"
	case iris.StatusUnauthorized:
		return "unauthorized"
	case iris.StatusForbidden:
		return "forbidden"
	case iris.StatusNotFound:
		return "not_found"
	case iris.StatusTooManyRequests:
		return "rate_limited"
	case iris.StatusInternalServerError:
		return "internal"
	default:
		return strings.ToLower(strings.ReplaceAll(http.StatusText(status), " ", "_"))
	}
}
