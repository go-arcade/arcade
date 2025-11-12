package repo

import (
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/google/wire"
)

// ProviderSet 提供仓储层相关的依赖
var ProviderSet = wire.NewSet(
	ProvideRepositories,
)

// ProvideRepositories 提供统一的 Repositories 实例
func ProvideRepositories(db database.DB, mongo database.MongoDB, cache cache.Cache) *Repositories {
	return NewRepositories(db, mongo, cache)
}
