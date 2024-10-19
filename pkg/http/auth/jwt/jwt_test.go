package jwt

import (
	"github.com/go-arcade/arcade/pkg/http"
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

func TestRefreshToken(t *testing.T) {

	userId := "1"
	secretKey := "bf284d03-ba65-42d4-a9fe-0d2fbfe61060"
	accessExpire := 3600 * time.Second
	refreshExpire := 7200 * time.Second
	aToken, rToken, err := GenToken(userId, []byte(secretKey), accessExpire, refreshExpire)
	if err != nil {
		t.Errorf("GenToken error: %v", err)
	}
	t.Logf("aToken: %s\n rToken: %s", aToken, rToken)

	auth := &http.Auth{
		SecretKey: secretKey,
	}
	newRefreshToken, err := RefreshToken(auth, userId, rToken)
	if err != nil {
		t.Errorf("RefreshToken error: %v", err)
	}
	t.Logf("newRefeshToken: %s", newRefreshToken)
}
