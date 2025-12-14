package middleware

import (
	"context"

	"github.com/go-arcade/arcade/pkg/trace"
	"github.com/go-arcade/arcade/pkg/trace/inject"
	"github.com/gofiber/fiber/v2"
)

// TraceMiddleware 链路追踪中间件
// 对 HTTP 服务器请求进行埋点
func TraceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 从 fiber context 中获取或创建 context
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}

		// 从 context 中获取 trace context（如果存在）
		ctx = trace.ContextWithSpan(ctx)

		// 使用埋点方法包装请求处理
		_, err := inject.HTTPServerRequest(ctx, c.Method(), c.Path(), func(ctx context.Context) (int, error) {
			// HTTPServerRequest 已经设置了 context，这里确保 context 也设置到 fiber context
			// 将 trace context 设置回 fiber context，以便后续中间件和处理器使用
			c.SetUserContext(ctx)

			// 处理请求
			nextErr := c.Next()

			// 返回状态码和错误
			return c.Response().StatusCode(), nextErr
		})

		// HTTPServerRequest 返回后，确保 context 仍然在 goroutine context 中
		// 这样 AccessLogFormat 中间件就能获取到 trace 信息
		// 注意：HTTPServerRequest 已经设置了 context，这里不需要再次设置

		// 如果埋点方法返回了错误，返回该错误
		return err
	}
}
