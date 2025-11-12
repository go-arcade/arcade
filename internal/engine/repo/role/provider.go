package role

import (
	"github.com/google/wire"
)

// ProviderSet 提供 role 模块的 repository 依赖
var ProviderSet = wire.NewSet(
	NewRoleRepo,
)
