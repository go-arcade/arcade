package interceptor

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 16:15
 * @file: cors_interceptor.go
 * @description: cors interceptor
 */

func CorsInterceptor() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 允许跨域访问的协议域名端口号
		AllowOrigins: []string{"*", "*"},
		// AllowOriginFunc优先级大于AllowOrigins
		// AllowOriginFunc: func(origin string) bool {
		// 	fmt.Println("origin:", origin)
		// 	// 允许跨域访问名单
		// 	var allowOriginsList = []string{"http://localhost:5173", "http://localhost:9999"}
		// 	//如果allowOriginsList中包含origin,允许访问
		// 	for _, v := range allowOriginsList {
		// 		// 如果访问域名是名单内的则放行
		// 		if strings.Contains(v, origin) {
		// 			return true
		// 		}
		// 	}
		// 	// 匹配不到不放行
		// 	return false
		// },
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		ExposeHeaders:    []string{"Content-Length"},
	})
}
