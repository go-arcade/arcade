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

func AccessLogFormat(log *zap.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {

		// exclude api path
		// tips: 这里的路径是不需要记录日志的路径，url为端口后的全部路径
		excludedPaths := map[string]bool{
			"/health": true,
		}

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		correctedLogger := log.WithOptions(zap.AddCallerSkip(-1), zap.AddCaller())

		c.Next()

		if excludedPaths[path] {
			c.Next()
			return
		}

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
