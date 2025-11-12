package agent

import (
	"github.com/google/wire"
)

// ProviderSet 提供 agent 模块的 repository 依赖
var ProviderSet = wire.NewSet(
	NewAgentRepo,
)
