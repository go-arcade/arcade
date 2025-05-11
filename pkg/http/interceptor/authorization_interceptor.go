package interceptor

import (
	"context"
	"errors"
	"strings"

	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/jwt"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 18:59
 * @file: authorization_interceptor.go
 * @description: authorization interceptor
 */

// AuthorizationInterceptor 鉴权拦截器
// This function is used as the middleware of fiber.
func AuthorizationInterceptor(secretKey, tokenPrefix string, client redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		aToken := c.Get("Authorization")
		if aToken == "" {
			http.WithRepErrMsg(c, http.TokenBeEmpty.Code, http.TokenBeEmpty.Msg, c.Path())
			return fiber.ErrUnauthorized
		}

		// 按空格分割
		parts := strings.SplitN(aToken, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			http.WithRepErrMsg(c, http.TokenBeEmpty.Code, http.TokenBeEmpty.Msg, c.Path())
			return fiber.ErrUnauthorized
		}

		claims, err := jwt.ParseToken(parts[1], secretKey)
		if err != nil {
			// 检查是否是令牌过期错误
			if errors.Is(err, goJwt.ErrTokenExpired) {
				http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Path())
				return fiber.ErrUnauthorized
			}
			// 其他令牌无效的情况
			http.WithRepErrMsg(c, http.InvalidToken.Code, http.InvalidToken.Msg, c.Path())
			log.Errorf("parse token failed: %v", err)
			return fiber.ErrUnauthorized
		}

		// 检查 Redis 中是否存在 Token
		tokenKey := tokenPrefix + claims.UserId
		exists, err := client.Exists(context.Background(), tokenKey).Result()
		if err != nil {
			http.WithRepErrMsg(c, http.InternalError.Code, http.InternalError.Msg, c.Path())
			log.Errorf("redis check token exists failed: %v", err)
			return fiber.ErrInternalServerError
		}
		if exists == 0 {
			http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Path())
			return fiber.ErrUnauthorized
		}

		// 检查 Redis 中的 Token 是否过期
		ttl, err := client.TTL(context.Background(), tokenKey).Result()
		if err != nil {
			http.WithRepErrMsg(c, http.InternalError.Code, http.InternalError.Msg, c.Path())
			log.Errorf("redis check token TTL failed: %v", err)
			return fiber.ErrInternalServerError
		}
		if ttl <= 0 {
			http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Path())
			log.Warnf("token has expired in Redis for user: %s", claims.UserId)
			return fiber.ErrUnauthorized
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}
