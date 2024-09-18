package interceptor

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/pkg/httpx"
	"net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/17 20:44
 * @file: unified_response.go
 * @description: 统一响应拦截器
 */

func UnifiedResponseInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 业务逻辑错误
		if len(c.Errors) > 0 && c.Writer.Status() != http.StatusOK {
			httpx.WithRepErrMsg(c, httpx.Failed.Code, httpx.Failed.Msg, c.Request.URL.Path)
			return
		}
		c.Next()

		if c.Writer.Status() == 0 {
			c.Writer.WriteHeader(httpx.Success.Code)
		}
		c.Next()

		// 业务逻辑正确
		if c.Writer.Status() >= http.StatusOK && c.Writer.Status() < http.StatusMultipleChoices {
			detail, exists := c.Get("detail")
			if exists && detail != nil {
				httpx.WithRepDetail(c, c.Writer.Status(), httpx.Success.Msg, detail)
				return
			}
		}
		c.Next()
	}
}
