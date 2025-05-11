package gprc

import (
	"sync"
)

var (
	mu      sync.RWMutex
	clients = make(map[string]any) // 存储所有注册的 Kitex 客户端
)

// Register 注册泛型客户端
func Register[T any](service string, cli T) {
	mu.Lock()
	defer mu.Unlock()
	clients[service] = cli
}

// Get 获取泛型客户端（必须提前注册）
func Get[T any](service string) T {
	mu.RLock()
	defer mu.RUnlock()
	return clients[service].(T)
}
