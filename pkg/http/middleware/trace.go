package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TraceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceId := c.Get("X-Request-Id")
		if traceId == "" {
			traceId = uuid.New().String()
		}
		c.Set("X-Request-Id", traceId)
		return c.Next()
	}
}
