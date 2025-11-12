package storage

import (
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/google/wire"
)

// ProviderSet 提供存储层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideStorageFromDB,
)

// ProvideStorageFromDB 从数据库提供存储提供者
func ProvideStorageFromDB(appCtx *ctx.Context, repos *repo.Repositories) (StorageProvider, error) {
	dbProvider, err := NewStorageDBProvider(appCtx, repos.Storage)
	if err != nil {
		return nil, err
	}
	return dbProvider.GetStorageProvider()
}

