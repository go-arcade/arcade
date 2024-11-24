package interceptor

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/internal/engine/tool"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/jwt"
	"github.com/go-arcade/arcade/pkg/log"
	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
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
func AuthorizationInterceptor(secretKey, tokenPrefix string, client redis.Client) gin.HandlerFunc {
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
			if errors.Is(err, goJwt.ErrTokenExpired) {
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

		token, err := tool.ParseAuthorizationToken(c, secretKey)
		if err != nil {
			return
		}

		isTokenExist(c, client, tokenPrefix+token.UserId)

		c.Set("claims", claims)
		c.Next()
	}
}

// isTokenExist 检查 Token 是否存在
func isTokenExist(c *gin.Context, client redis.Client, token string) {
	exists, err := client.Exists(context.Background(), token).Result()
	if err != nil {
		// Redis 出错
		http.WithRepErrMsg(c, http.InternalError.Code, http.InternalError.Msg, c.Request.URL.Path)
		log.Errorf("redis check token exists failed: %v", err)
		c.Abort()
		return
	}
	if exists == 0 {
		// Token 不存在
		http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Request.URL.Path)
		c.Abort()
		return
	}
}
