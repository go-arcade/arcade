package http

import (
	"time"

	"github.com/go-arcade/arcade/pkg/log"
	tracecontext "github.com/go-arcade/arcade/pkg/trace/context"
	"github.com/gofiber/fiber/v2"
)

// AccessLogFormat 创建访问日志中间件（动态检查配置）
func AccessLogFormat(httpConfig *Http) fiber.Handler {
	// exclude api path
	// tips: 这里的路径是不需要记录日志的路径，url为端口后的全部路径
	excludedPaths := map[string]bool{
		"/health": true,
	}

	return func(c *fiber.Ctx) error {
		if httpConfig != nil && !httpConfig.AccessLog {
			return c.Next()
		}

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

		// 确保 trace context 设置到 goroutine context，以便日志能获取到 trace 信息
		// trace middleware 应该已经设置了 context，但这里再次确保
		// 从 fiber context 中获取 trace context，并设置到 goroutine context
		if ctx := c.UserContext(); ctx != nil {
			// 确保 context 中包含有效的 span
			tracecontext.SetContext(ctx)
		}
		// 在记录日志后清除 context
		defer tracecontext.ClearContext()

		log.Infow("HTTP request",
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
