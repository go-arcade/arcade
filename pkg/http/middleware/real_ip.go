package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RealIPMiddleware 获取真实 IP 中间件
func RealIPMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if xff := c.Get("X-Forwarded-For"); xff != "" {
			// XFF: client, proxy1, proxy2
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				ip := strings.TrimSpace(parts[0])
				if ip != "" {
					c.Locals("ip", ip)
					return c.Next()
				}
			}
		}
		if ip := c.Get("X-Real-IP"); ip != "" {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				c.Locals("ip", ip)
			}
		}
		return c.Next()
	}
}
