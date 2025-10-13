package repo

import (
	"github.com/google/wire"
	"github.com/observabil/arcade/pkg/ctx"
)

// ProviderSet 提供仓储层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideAgentRepo,
	ProvideUserRepo,
	ProvidePluginRepo,
	ProvidePluginRepoAdapter,
	ProvideSSORepo,
)

// ProvideAgentRepo 提供 Agent 仓储实例
func ProvideAgentRepo(ctx *ctx.Context) *AgentRepo {
	return NewAgentRepo(ctx)
}

// ProvideUserRepo 提供 User 仓储实例
func ProvideUserRepo(ctx *ctx.Context) *UserRepo {
	return NewUserRepo(ctx)
}

// ProvidePluginRepo 提供 Plugin 仓储实例
func ProvidePluginRepo(ctx *ctx.Context) *PluginRepo {
	return NewPluginRepo(ctx)
}

// ProvidePluginRepoAdapter 提供 PluginRepoAdapter 实例
func ProvidePluginRepoAdapter(repo *PluginRepo) *PluginRepoAdapter {
	return NewPluginRepoAdapter(repo)
}

// ProvideSSORepo 提供 SSO 仓储实例
func ProvideSSORepo(ctx *ctx.Context) *SSORepo {
	return NewSSORepo(ctx)
}
