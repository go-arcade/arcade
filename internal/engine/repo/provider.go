package repo

import (
	"github.com/google/wire"
	"github.com/observabil/arcade/pkg/ctx"
)

// ProviderSet 提供仓储层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideAgentRepo,
	ProvideUserRepo,
)

// ProvideAgentRepo 提供 Agent 仓储实例
func ProvideAgentRepo(ctx *ctx.Context) *AgentRepo {
	return NewAgentRepo(ctx)
}

// ProvideUserRepo 提供 User 仓储实例
func ProvideUserRepo(ctx *ctx.Context) *UserRepo {
	return NewUserRepo(ctx)
}
