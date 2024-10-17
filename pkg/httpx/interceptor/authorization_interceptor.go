package interceptor

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/pkg/httpx"
	"github.com/go-arcade/arcade/pkg/httpx/auth/jwt"
	"github.com/go-arcade/arcade/pkg/log"
	"net/http"
	"strings"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 18:59
 * @file: authorization_interceptor.go
 * @description: authorization interceptor
 */

// AuthorizationInterceptor 鉴权拦截器
// This function is used as the middleware of gin.
func AuthorizationInterceptor(accessExpired, refreshExpired time.Duration, secretKey string) gin.HandlerFunc {
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

		secret, exists := c.Get("secretKey")
		if !exists {
			log.Error("secretKey not exists")
			c.Abort()
			return
		}
		mc, err := jwt.ParseToken(parts[1], secret.(string))
		if err != nil {
			httpx.WithRepErrMsg(c, httpx.TokenInvalid.Code, httpx.Unauthorized.Msg, c.Request.URL.Path)
			c.Abort()
			return
		}

		aToken, rToken, err := jwt.RefreshToken(parts[1], parts[2])
		if err != nil {
			httpx.WithRepErr(c, http.StatusUnauthorized, httpx.Unauthorized.Msg, c.Request.URL.Path)
			c.Abort()
			return
		} else {
			token := make(map[string]string)
			token["accessToken"] = aToken
			token["refreshToken"] = rToken
			c.Set("details", token)
		}

		c.Set("claims", mc)
		c.Next()
	}
}
