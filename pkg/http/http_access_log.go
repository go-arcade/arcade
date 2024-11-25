package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 23:01
 * @file: http_access_log.go
 * @description:
 */

func AccessLogFormat(log zap.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		correctedLogger := log.WithOptions(zap.AddCallerSkip(1))

		c.Next()

		cost := time.Since(start).Seconds() // 转换为秒
		correctedLogger.Sugar().Debugf(
			"[%s %s%s] %s %s %d %.2fs",
			c.Request.Method,      // HTTP 方法
			path,                  // 请求路径
			formatQuery(query),    // 查询参数
			c.ClientIP(),          // 客户端 IP
			c.Request.UserAgent(), // User-Agent
			c.Writer.Status(),     // 响应状态码
			cost,                  // 请求耗时（秒）
		)
	}
}

func formatQuery(query string) string {
	if query != "" {
		return "?" + query
	}
	return ""
}
