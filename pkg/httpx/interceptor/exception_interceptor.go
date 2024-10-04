package interceptor

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/pkg/httpx"
	"github.com/go-arcade/arcade/pkg/log"
	"runtime/debug"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 1:07
 * @file: exception_interceptor.go
 * @description:
 */

// ExceptionInterceptor 异常拦截器
func ExceptionInterceptor(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			httpx.WithRepErr(c, httpx.InternalError.Code, errorToString(err), c.Request.URL.Path)
			//log.Errorf("panic: %v", err)
			c.Abort()
		}
	}()
	c.Next()

}

func errorToString(err interface{}) string {
	switch v := err.(type) {
	case httpx.ResponseErr:
		// 符合预期的错误，可以直接返回给客户端
		if errMsg, ok := v.ErrMsg.(string); ok {
			return errMsg
		}
		// 如果 ErrMsg 不是字符串类型，则返回默认错误消息
		return httpx.InternalError.Msg
	case error:
		// 一律返回服务器错误，避免返回堆栈错误给客户端，实际还可以针对系统错误做其他处理
		debug.PrintStack()
		log.Errorf("panic: %v\n%s", v, debug.Stack())
		return httpx.InternalError.Msg
	default:
		if errMsg, ok := v.(string); ok {
			return errMsg
		}
		return httpx.InternalError.Msg
	}
}
