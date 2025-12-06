package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func AccessLogFormat(log *zap.Logger) fiber.Handler {
	// 使用 sugar logger
	sugar := log.Sugar()
	// exclude api path
	// tips: 这里的路径是不需要记录日志的路径，url为端口后的全部路径
	excludedPaths := map[string]bool{
		"/health": true,
	}

	return func(c *fiber.Ctx) error {
		// 检查是否需要跳过日志
		if excludedPaths[c.Path()] {
			return c.Next()
		}

		// 记录开始时间
		start := time.Now()

		// 处理请求
		err := c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 构建查询字符串
		query := c.Context().QueryArgs().String()
		queryStr := ""
		if query != "" {
			queryStr = "?" + query
		}

		// 使用 sugar logger 记录访问日志
		sugar.Infow("HTTP request",
			"method", c.Method(),
			"path", c.Path(),
			"query", queryStr,
			"status", c.Response().StatusCode(),
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
			"latency", latency.String(),
		)

		return err
	}
}
