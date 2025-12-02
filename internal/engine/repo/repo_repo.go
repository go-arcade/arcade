package repo

import (
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"gorm.io/gorm"
)

// Repositories 统一管理所有 repository
type Repositories struct {
	User                IUserRepository
	Agent               IAgentRepository
	Plugin              IPluginRepository
	PluginTask          IPluginTaskRepository
	Permission          IPermissionRepository
	Role                IRoleRepository
	Storage             IStorageRepository
	Team                ITeamRepository
	IdentityIntegration IIdentityIntegrationRepository
	GeneralSettings     IGeneralSettingsRepository
	RouterPermission    IRouterPermissionRepository
	ProjectMember       IProjectMemberRepository
	ProjectTeamAccess   IProjectTeamAccessRepository
	TeamMember          ITeamMemberRepository
	UserExtension       IUserExtensionRepository
	Secret              ISecretRepository
}

// NewRepositories 初始化所有 repository
func NewRepositories(db database.IDatabase, mongo database.MongoDB, cache cache.ICache) *Repositories {
	return &Repositories{
		User:                NewUserRepo(db, cache),
		Agent:               NewAgentRepo(db),
		Plugin:              NewPluginRepo(db),
		PluginTask:          NewPluginTaskRepo(mongo),
		Permission:          NewPermissionRepo(db, cache),
		Role:                NewRoleRepo(db),
		Storage:             NewStorageRepo(db, cache),
		Team:                NewTeamRepo(db),
		IdentityIntegration: NewIdentityIntegrationRepo(db),
		GeneralSettings:     NewGeneralSettingsRepo(db),
		RouterPermission:    NewRouterPermissionRepo(db),
		ProjectMember:       NewProjectMemberRepo(db),
		ProjectTeamAccess:   NewProjectTeamAccessRepo(db),
		TeamMember:          NewTeamMemberRepo(db),
		UserExtension:       NewUserExtensionRepo(db),
		Secret:              NewSecretRepo(db),
	}
}

// GetDB 返回数据库实例（供插件适配器使用）
func (r *Repositories) GetDB() database.IDatabase {
	return r.Storage.(*StorageRepo).GetDB()
}

func Count(tx *gorm.DB) (int64, error) {
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func Exist(tx *gorm.DB, where interface{}) bool {
	var one interface{}
	if err := tx.Where(where).First(&one).Error; err != nil {
		return false
	}
	return true
}
