package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestMiddleware set request id
func RequestMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestId := c.Request().Header.Peek("X-Request-Id")
		if len(requestId) == 0 {
			requestId = []byte(uuid.New().String())
		}
		c.Request().Header.Set("X-Request-Id", string(requestId))
		c.Set("X-Request-Id", string(requestId))
		c.Locals("request_id", string(requestId))
		return c.Next()
	}
}
