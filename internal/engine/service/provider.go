package service

import (
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/service/agent"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/google/wire"
)

// ProviderSet 提供服务层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideAgentService,
	ProvideUserService,
)

// ProvideAgentService 提供 Agent 服务实例
func ProvideAgentService(agentRepo *repo.AgentRepo) *agent.AgentService {
	return &agent.AgentService{
		AgentRepo: agentRepo,
	}
}

// ProvideUserService 提供 User 服务实例
func ProvideUserService(ctx *ctx.Context, userRepo *repo.UserRepo) *UserService {
	return NewUserService(ctx, userRepo)
}
