package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/go-arcade/arcade/internal/engine/consts"
	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/jwt"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// AuthorizationMiddleware 认证中间件
// secretKey: 用于验证 JWT 的密钥
// client: Redis 客户端
// This function is used as the middleware of fiber.
func AuthorizationMiddleware(secretKey string, client redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		aToken := c.Get("Authorization")
		if aToken == "" {
			return http.WithRepErrMsg(c, http.TokenBeEmpty.Code, http.TokenBeEmpty.Msg, c.Path())
		}

		// 按空格分割
		parts := strings.SplitN(aToken, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return http.WithRepErrMsg(c, http.TokenBeEmpty.Code, http.TokenBeEmpty.Msg, c.Path())
		}

		claims, err := jwt.ParseToken(parts[1], secretKey)
		if err != nil {
			// 检查是否是令牌过期错误
			if errors.Is(err, goJwt.ErrTokenExpired) {
				return http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Path())
			}
			log.Errorf("parse token failed: %v", err)
			// 其他令牌无效的情况
			return http.WithRepErrMsg(c, http.InvalidToken.Code, http.InvalidToken.Msg, c.Path())
		}

		// 检查 Redis 中是否存在 Token
		tokenKey := consts.UserInfoKey + claims.UserId
		exists, err := client.Exists(context.Background(), tokenKey).Result()
		if err != nil {
			log.Errorf("redis check token exists failed: %v", err)
			return http.WithRepErrMsg(c, http.InternalError.Code, http.InternalError.Msg, c.Path())
		}
		if exists == 0 {
			return http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Path())
		}

		// 检查 Redis 中的 Token 是否过期
		ttl, err := client.TTL(context.Background(), tokenKey).Result()
		if err != nil {
			log.Errorf("redis check token TTL failed: %v", err)
			return http.WithRepErrMsg(c, http.InternalError.Code, http.InternalError.Msg, c.Path())
		}
		if ttl <= 0 {
			log.Warnf("token has expired in Redis for user: %s", claims.UserId)
			return http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Path())
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}
