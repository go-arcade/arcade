package service

import (
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"github.com/google/wire"
)

// ProviderSet 提供服务层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideServices,
)

// ProvideServices 提供统一的 Services 实例
func ProvideServices(
	ctx *ctx.Context,
	db database.DB,
	cache cache.Cache,
	repos *repo.Repositories,
	pluginManager *pluginpkg.Manager,
	storageProvider storage.StorageProvider,
) *Services {
	return NewServices(ctx, db, cache, repos, pluginManager, storageProvider)
}
