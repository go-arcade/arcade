package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var (
	allowMethods  = "GET, POST, PUT, DELETE, OPTIONS"
	allowHeaders  = "Origin, X-Requested-With, Content-Type, Accept, Authorization"
	exposeHeaders = "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type"
)

// CorsMiddleware 跨域中间件
func CorsMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: true,
	})
}
