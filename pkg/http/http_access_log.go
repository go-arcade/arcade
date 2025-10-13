package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 23:01
 * @file: http_access_log.go
 * @description:
 */

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
		sugar.Infof("[%s %s %s %d] %s %s %s",
			c.Method(),
			c.Path(),
			queryStr,
			c.Response().StatusCode(),
			c.IP(),
			c.Get("User-Agent"),
			latency.String(),
		)

		return err
	}
}
