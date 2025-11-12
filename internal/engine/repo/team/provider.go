package team

import (
	"github.com/google/wire"
)

// ProviderSet 提供 team 模块的 repository 依赖
var ProviderSet = wire.NewSet(
	NewTeamRepo,
	NewTeamMemberRepo,
)
