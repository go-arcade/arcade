package middleware

import (
	"runtime/debug"

	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
)

// ExceptionMiddleware 异常中间件
// 捕获 panic 错误，返回 500 状态码和错误信息
// This function is used as the middleware of fiber.
func ExceptionMiddleware(c *fiber.Ctx) error {
	defer func() {
		if err := recover(); err != nil {
			_ = http.WithRepErr(c, http.InternalError.Code, errorToString(err), c.Path())
			log.Errorf("panic: %v", err)
		}
	}()

	return c.Next()
}

func errorToString(err any) string {
	switch v := err.(type) {
	case http.ResponseErr:
		// 符合预期的错误，可以直接返回给客户端
		if errMsg, ok := v.ErrMsg.(string); ok {
			return errMsg
		}
		// 如果 ErrMsg 不是字符串类型，则返回默认错误消息
		return http.InternalError.Msg
	case error:
		// 一律返回服务器错误，避免返回堆栈错误给客户端，实际还可以针对系统错误做其他处理
		debug.PrintStack()
		log.Errorf("panic: %v\n%s", v, debug.Stack())
		return http.InternalError.Msg
	default:
		if errMsg, ok := v.(string); ok {
			return errMsg
		}
		return http.InternalError.Msg
	}
}
