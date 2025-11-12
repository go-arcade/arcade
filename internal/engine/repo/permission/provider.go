package permission

import (
	"github.com/google/wire"
)

// ProviderSet 提供 permission 模块的 repository 依赖
var ProviderSet = wire.NewSet(
	NewPermissionRepo,
	NewRouterPermissionRepo,
)
