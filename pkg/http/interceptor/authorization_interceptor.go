package interceptor

import (
	"github.com/gin-gonic/gin"
	httpx "github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/auth/jwt"
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
func AuthorizationInterceptor(secretKey string, auth httpx.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		aToken := c.Request.Header.Get("Authorization")
		if aToken == "" {
			httpx.WithRepErrMsg(c, httpx.TokenBeEmpty.Code, httpx.TokenBeEmpty.Msg, c.Request.URL.Path)
			c.Abort()
			return
		}

		// 按空格分割
		parts := strings.SplitN(aToken, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			httpx.WithRepErrMsg(c, httpx.TokenBeEmpty.Code, httpx.TokenBeEmpty.Msg, c.Request.URL.Path)
			c.Abort()
			return
		}

		refreshToken := c.Request.Header.Get("X-Refresh-Token")
		claims, err := jwt.ParseToken(aToken, secretKey)
		if err != nil {
			// 访问令牌无效或过期，尝试刷新令牌
			if refreshToken == "" {
				// 没有刷新令牌，无法刷新
				httpx.WithRepErrMsg(c, httpx.InvalidToken.Code, httpx.InvalidToken.Msg, c.Request.URL.Path)
				c.Abort()
				return
			}

			userId := claims.UserId
			if userId == "" {
				httpx.WithRepErrMsg(c, httpx.InvalidToken.Code, httpx.UserNotExist.Msg, c.Request.URL.Path)
				c.Abort()
				c.Set("userId", userId)
				return
			}

			// 尝试刷新令牌
			newTokens, err := jwt.RefreshToken(&auth, userId, refreshToken)
			if err != nil {
				// 刷新令牌无效或过期
				httpx.WithRepErr(c, httpx.Unauthorized.Code, httpx.Unauthorized.Msg, c.Request.URL.Path)
				c.Abort()
				return
			} else {
				// 成功刷新令牌，将其设置到响应或上下文中
				c.Set("details", newTokens)
				// 解析新的访问令牌以获取声明
				claims, err = jwt.ParseToken(newTokens["accessToken"], secretKey)
				if err != nil {
					httpx.WithRepErr(c, httpx.InvalidToken.Code, httpx.InvalidToken.Msg, c.Request.URL.Path)
					c.Abort()
					return
				}
			}
		}

		c.Set("claims", claims)
		c.Next()
	}
}
