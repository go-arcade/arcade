package id

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/16 21:53
 * @file: ulid.go
 * @description: ulid
 */

func GetUild() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)
	if err != nil {
		return ""
	}
	return id.String()
}
