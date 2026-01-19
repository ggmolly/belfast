package middleware

import (
	"time"

	"github.com/kataras/iris/v12"

	"github.com/ggmolly/belfast/internal/logger"
)

func RequestLogger() iris.Handler {
	return func(ctx iris.Context) {
		start := time.Now()
		ctx.Next()
		latency := time.Since(start)
		logger.WithFields(
			"API",
			logger.FieldValue("method", ctx.Method()),
			logger.FieldValue("path", ctx.Path()),
			logger.FieldValue("status", ctx.GetStatusCode()),
			logger.FieldValue("latency", latency.String()),
			logger.FieldValue("ip", ctx.RemoteAddr()),
		).Info("request")
	}
}
