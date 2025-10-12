package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/observabil/arcade/pkg/http"
	"github.com/observabil/arcade/pkg/log"
)

/**
 * @author: x.gallagher.anderson@gmail.com
 * @time: 2023/11/14 22:40
 * @file: jwt.go
 * @description:
 */

type AuthClaims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

func (a *AuthClaims) Valid() error {
	return nil
}

var (
	issUser = "arcade"
)

func keyFunc(auth http.Auth) (any, error) {
	return auth.SecretKey, nil
}

// GenToken 生成 access_token 和 refresh_token
func GenToken(userId string, secretKey []byte, accessExpired, refreshExpired time.Duration) (aToken, rToken string, err error) {

	// aToken
	aClaims := &AuthClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issUser, // 签发人
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpired * time.Minute)),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	aToken, aErr := jwt.NewWithClaims(jwt.SigningMethodHS256, aClaims).SignedString(secretKey)
	if aErr != nil {
		log.Errorf("jwt.NewWithClaims err: %v", aErr)
		return "", "", aErr
	}

	// rToken
	rClaims := jwt.RegisteredClaims{
		Issuer:    issUser,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpired * time.Minute)),
	}
	rToken, rErr := jwt.NewWithClaims(jwt.SigningMethodHS256, rClaims).SignedString(secretKey)
	if rErr != nil {
		log.Debugf("jwt.NewWithClaims err: %v", rErr)
		return "", "", rErr
	}

	return aToken, rToken, nil
}

// ParseToken 校验 access_token
func ParseToken(aToken, secretKey string) (claims *AuthClaims, err error) {
	claims = new(AuthClaims)
	token, err := jwt.ParseWithClaims(aToken, claims, func(token *jwt.Token) (any, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		// 细化错误处理
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken 刷新 access_token
func RefreshToken(auth *http.Auth, userId, rToken string) (map[string]string, error) {
	newToken := make(map[string]string)

	// 解析刷新令牌
	var refreshClaims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(rToken, &refreshClaims, func(token *jwt.Token) (any, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(auth.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return newToken, errors.New(http.InvalidToken.Msg)
	}

	// 检查刷新令牌是否过期
	if refreshClaims.ExpiresAt.Before(time.Now()) {
		return newToken, errors.New(http.TokenExpired.Msg)
	}

	// 生成新的访问令牌和刷新令牌
	newAToken, newRToken, err := GenToken(userId, []byte(auth.SecretKey), auth.AccessExpire, auth.RefreshExpire)
	if err != nil {
		return newToken, err
	}

	newToken["accessToken"] = newAToken
	newToken["refreshToken"] = newRToken

	return newToken, nil
}
