package interceptor

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 16:15
 * @file: cors_interceptor.go
 * @description: cors interceptor
 */

func CorsInterceptor() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, X-Requested-With, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type",
		AllowCredentials: true,
	})
}
