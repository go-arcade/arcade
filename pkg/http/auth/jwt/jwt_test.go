package jwt

import (
	"testing"
	"time"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/10/17 22:47
 * @file: jwt_test.go
 * @description:
 */

func TestJwt(t *testing.T) {

	userId := "1"
	secretKey := []byte("1111111111111111")
	accessExpired := time.Hour * 24
	refreshExpired := time.Hour * 24 * 7

	aToken, rToken, err := GenToken(userId, secretKey, accessExpired, refreshExpired)
	if err != nil {
		t.Errorf("GenToken error: %v", err)
	}
	t.Logf("aToken: %s, rToken: %s", aToken, rToken)
}
