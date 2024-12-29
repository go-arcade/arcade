package interceptor

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/engine/constant"
	httpx "github.com/go-arcade/arcade/pkg/http"
	"net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/17 20:44
 * @file: unified_resp_interceptor.go
 * @description: 统一响应拦截器
 */

// UnifiedResponseInterceptor 统一响应拦截器
// c.Set("detail", value) 用于设置响应数据
// 如有其他需要，可自行添加
func UnifiedResponseInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 如果已经被中止，说明其他中间件已经处理了响应
		if c.IsAborted() {
			return
		}

		// 业务逻辑错误
		if len(c.Errors) > 0 && c.Writer.Status() != http.StatusOK {
			httpx.WithRepErrMsg(c, httpx.Failed.Code, httpx.Failed.Msg, c.Request.URL.Path)
			return
		}

		// 如果未设置响应状态码，默认将状态码设置为200（OK）
		if c.Writer.Status() == 0 {
			c.Writer.WriteHeader(httpx.Success.Code)
		}

		// 业务逻辑正确, 设置响应数据
		if c.Writer.Status() >= http.StatusOK && c.Writer.Status() < http.StatusMultipleChoices {
			if detail, exists := c.Get(constant.DETAIL); exists && detail != nil {
				httpx.WithRepJSON(c, detail)
				return
			}

			// 业务逻辑正确, 无响应数据, 只返回结果
			if _, exists := c.Get(constant.OPERATION); exists {
				httpx.WithRepNotDetail(c)
				return
			}
		}
	}
}
