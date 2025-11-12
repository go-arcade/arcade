package plugin

import (
	"github.com/google/wire"
)

// ProviderSet 提供 plugin 模块的 repository 依赖
var ProviderSet = wire.NewSet(
	NewPluginRepo,
	NewPluginTaskRepo,
)
