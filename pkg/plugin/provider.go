package plugin

import (
	"github.com/google/wire"
)

// ProviderSet 提供插件相关的依赖
var ProviderSet = wire.NewSet(ProvidePluginManager)

// ProvidePluginManager 提供插件管理器实例
func ProvidePluginManager() *Manager {
	return NewManager()
}
