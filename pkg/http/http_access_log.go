package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 23:01
 * @file: http_access_log.go
 * @description:
 */

func AccessLogFormat(log *zap.Logger) fiber.Handler {
	// exclude api path
	// tips: 这里的路径是不需要记录日志的路径，url为端口后的全部路径
	excludedPaths := map[string]bool{
		"/health": true,
	}

	return logger.New(logger.Config{
		Format:     "[${method} ${path}${query}] ${ip} ${userAgent} ${status} ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		CustomTags: map[string]logger.LogFunc{
			"query": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				query := c.Context().QueryArgs().String()
				if query != "" {
					return output.WriteString("?" + query)
				}
				return 0, nil
			},
		},
		Next: func(c *fiber.Ctx) bool {
			return excludedPaths[c.Path()]
		},
	})
}
