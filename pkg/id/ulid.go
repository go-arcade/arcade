package id

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)


func GetUild() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)
	if err != nil {
		return ""
	}
	return id.String()
}
