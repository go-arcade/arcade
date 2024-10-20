package interceptor

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/auth/jwt"
	"github.com/go-arcade/arcade/pkg/log"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"strings"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 18:59
 * @file: authorization_interceptor.go
 * @description: authorization interceptor
 */

// AuthorizationInterceptor 鉴权拦截器
// This function is used as the middleware of gin.
func AuthorizationInterceptor(secretKey string, auth http.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		aToken := c.Request.Header.Get("Authorization")
		if aToken == "" {
			http.WithRepErrMsg(c, http.TokenBeEmpty.Code, http.TokenBeEmpty.Msg, c.Request.URL.Path)
			c.Abort()
			return
		}

		// 按空格分割
		parts := strings.SplitN(aToken, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			http.WithRepErrMsg(c, http.TokenBeEmpty.Code, http.TokenBeEmpty.Msg, c.Request.URL.Path)
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1], secretKey)
		if err != nil {
			// 检查是否是令牌过期错误
			if errors.Is(err, jwt2.ErrTokenExpired) {
				http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Request.URL.Path)
				c.Abort()
				return
			}
			// 其他令牌无效的情况，返回错误
			http.WithRepErrMsg(c, http.InvalidToken.Code, http.InvalidToken.Msg, c.Request.URL.Path)
			log.Errorf("parse token failed: %v", err)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
