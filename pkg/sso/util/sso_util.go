package util

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

var StateStore = &sync.Map{}

// GenState generate a random state string
func GenState() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// LoadAndDeleteState load and delete the state from the state store
func LoadAndDeleteState(state string) (string, bool) {
	value, ok := StateStore.LoadAndDelete(state)
	if ok {
		return value.(string), true
	}
	return "", false
}
