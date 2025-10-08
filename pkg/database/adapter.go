package database

import (
	_ "embed"
	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
	"sync"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/28 21:39
 * @file: adapter.go
 * @description: casbin gorm adapter
 */

type CasbinService struct{}

var CasbinServiceApp = new(CasbinService)

var (
	syncedEnforcer *casbin.SyncedEnforcer
	once           sync.Once
)

func (c *CasbinService) Casbin(db *gorm.DB) *casbin.SyncedEnforcer {

	once.Do(func() {
		a, _ := casbin.NewSyncedEnforcer(db)
		syncedEnforcer, _ = casbin.NewSyncedEnforcer("conf.d/model.conf", a)
	})

	_ = syncedEnforcer.LoadPolicy()
	return syncedEnforcer
}
