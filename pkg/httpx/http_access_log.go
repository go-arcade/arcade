package httpx

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 23:01
 * @file: http_access_log.go
 * @description:
 */

func AccessLogFormat(param gin.LogFormatterParams) string {

	return fmt.Sprintf("%s %s [%s %s] %s %s %v %s %s\n",
		param.TimeStamp.Format("2006/01/02 15:04:05"), // 2024/09/16 22:57:40
		param.Latency,
		param.Method,
		param.Path,
		param.ClientIP,
		param.Request.Proto,
		param.StatusCode,
		param.Request.UserAgent(),
		param.ErrorMessage,
	)
}
