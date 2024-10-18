package jwt

import (
	"errors"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

/**
 * @author: x.gallagher.anderson@gmail.com
 * @time: 2023/11/14 22:40
 * @file: jwt.go
 * @description:
 */

type AuthClaims struct {
	Id string `json:"id"`
	//Name           string        `json:"name"`
	//AccessExpired  time.Duration `json:"accessExpired"`
	//RefreshExpired time.Duration `json:"refreshExpired"`
	//SecretKey      string        `json:"secretKey"`
	jwt.RegisteredClaims
}

func (a *AuthClaims) Valid() error {
	return nil
}

var (
	issUser = "arcade"
)

func keyFunc(token *jwt.Token) (interface{}, error) {
	return "secretKey", nil
}

// GenToken 生成 access_token 和 refresh_token
func GenToken(userId string, secretKey []byte, accessExpired, refreshExpired time.Duration) (aToken, rToken string, err error) {

	// aToken
	aClaims := &AuthClaims{
		Id: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issUser, // 签发人
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpired)),
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
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpired)),
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
	token, err := jwt.ParseWithClaims(aToken, &AuthClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		// 直接使用标准的Claim则可以直接使用Parse方法
		//token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	// 对token对象中的Claim进行类型断言
	if authClaims, ok := token.Claims.(*AuthClaims); ok && token.Valid { // 校验token
		return authClaims, nil
	}
	return nil, errors.New("invalid token")
}

// RefreshToken 刷新 access_token
func RefreshToken(aToken, rToken string) (aTokenNew, rTokenNew string, err error) {

	if _, err := jwt.Parse(rToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	}); err != nil {
		return "", "", err
	}

	var claims AuthClaims
	//var v *jwt.ClaimsValidator
	_, err = jwt.ParseWithClaims(aToken, &claims, keyFunc)
	if err != nil {
		return "", "", err
	}
	//errors.As(err, &v)
	//if v.Valid() != nil {
	//	return "", "", err
	//}

	return "", "", err
}
