package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/go-arcade/arcade/internal/engine/consts"
	"github.com/go-arcade/arcade/internal/engine/model"
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

		// 从 Redis 中获取 Token 信息
		tokenKey := consts.UserTokenKey + claims.UserId
		tokenInfoStr, err := client.Get(context.Background(), tokenKey).Result()
		if err != nil {
			log.Errorf("redis get token failed: %v", err)
			return http.WithRepErrMsg(c, http.TokenExpired.Code, http.TokenExpired.Msg, c.Path())
		}

		// 解析 Token 信息
		var tokenInfo model.TokenInfo
		if err := sonic.UnmarshalString(tokenInfoStr, &tokenInfo); err != nil {
			log.Errorf("failed to unmarshal token info: %v", err)
			return http.WithRepErrMsg(c, http.InvalidToken.Code, http.InvalidToken.Msg, c.Path())
		}

		// 验证请求中的 Token 是否与 Redis 中存储的 Token 匹配
		if tokenInfo.AccessToken != parts[1] {
			log.Warnf("token mismatch for user: %s", claims.UserId)
			return http.WithRepErrMsg(c, http.InvalidToken.Code, http.InvalidToken.Msg, c.Path())
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}
