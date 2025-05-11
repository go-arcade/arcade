package tool

import (
	"errors"
	"strings"

	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/http/jwt"
	"github.com/gofiber/fiber/v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 19:51
 * @file: token.go
 * @description: token tool
 */

func ParseAuthorizationToken(c *fiber.Ctx, secretKey string) (*jwt.AuthClaims, error) {
	token := c.Get("Authorization")
	if token == "" {
		return nil, errors.New(http.TokenBeEmpty.Msg)
	}
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	} else {
		// 处理令牌格式不正确的情况
		return nil, errors.New(http.TokenFormatIncorrect.Msg)
	}
	claims, err := jwt.ParseToken(token, secretKey)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
