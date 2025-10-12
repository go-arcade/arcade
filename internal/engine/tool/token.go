package tool

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/http/jwt"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/20 19:51
 * @file: token.go
 * @description: token tool
 */

// ParseAuthorizationToken 解析 Authorization 头中的 Bearer token
func ParseAuthorizationToken(f *fiber.Ctx, secretKey string) (*jwt.AuthClaims, error) {
	token := f.Get("Authorization")
	if token == "" {
		return nil, errors.New(http.TokenBeEmpty.Msg)
	}

	if t, ok := strings.CutPrefix(token, "Bearer "); ok {
		token = t
	} else {
		return nil, errors.New(http.TokenFormatIncorrect.Msg)
	}

	claims, err := jwt.ParseToken(token, secretKey)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
