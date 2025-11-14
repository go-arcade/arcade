package id

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)


var mu = &sync.Mutex{}

// GetUUID generates a new UUID
func GetUUID() string {
	mu.Lock()
	defer mu.Unlock()
	return uuid.NewString()
}

// GetUUIDWithoutDashes generates a new UUID not horizontal line
func GetUUIDWithoutDashes() string {
	mu.Lock()
	defer mu.Unlock()

	return strings.Replace(uuid.NewString(), "-", "", -1)
}
