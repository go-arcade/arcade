package interceptor

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/app/engine/consts"
	"github.com/go-arcade/arcade/pkg/httpx"
	"net/http"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/17 20:44
 * @file: unified_resp_interceptor_.go
 * @description: 统一响应拦截器
 */

// UnifiedResponseInterceptor 统一响应拦截器
// c.Set("detail", value) 用于设置响应数据
// 如有其他需要，可自行添加
func UnifiedResponseInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 业务逻辑错误
		if len(c.Errors) > 0 && c.Writer.Status() != http.StatusOK {
			httpx.WithRepErrMsg(c, httpx.Failed.Code, httpx.Failed.Msg, c.Request.URL.Path)
			return
		}
		c.Next()

		// 如果未设置响应状态码，默认将状态码设置为200（OK）
		if c.Writer.Status() == 0 {
			c.Writer.WriteHeader(httpx.Success.Code)
		}
		c.Next()

		// 业务逻辑正确, 设置响应数据
		if c.Writer.Status() >= http.StatusOK && c.Writer.Status() < http.StatusMultipleChoices {
			detail, exists := c.Get(consts.DETAIL)
			if exists && detail != nil {
				httpx.WithRepJSON(c, detail)
				return
			}
		}
		c.Next()

		// 业务逻辑正确, 无响应数据, 只返回结果
		if c.Writer.Status() >= http.StatusOK && c.Writer.Status() < http.StatusMultipleChoices {
			detail, exists := c.Get(consts.OPERATION)
			if exists && detail != nil {
				httpx.WithRepNotDetail(c)
				return
			}
		}
		c.Next()
	}
}
