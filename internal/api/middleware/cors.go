package middleware

import (
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

func CORS(origins []string) iris.Handler {
	if len(origins) == 0 {
		return func(ctx iris.Context) {
			ctx.Next()
		}
	}
	options := cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           600,
	}
	return cors.New(options)
}
