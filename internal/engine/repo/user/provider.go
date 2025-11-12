package user

import (
	"github.com/google/wire"
)

// ProviderSet 提供 user 模块的 repository 依赖
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewUserExtensionRepo,
)
