package repo

import (
	"github.com/go-arcade/arcade/internal/engine/repo/agent"
	"github.com/go-arcade/arcade/internal/engine/repo/general_settings"
	"github.com/go-arcade/arcade/internal/engine/repo/identity_integration"
	"github.com/go-arcade/arcade/internal/engine/repo/permission"
	"github.com/go-arcade/arcade/internal/engine/repo/plugin"
	"github.com/go-arcade/arcade/internal/engine/repo/project"
	"github.com/go-arcade/arcade/internal/engine/repo/role"
	"github.com/go-arcade/arcade/internal/engine/repo/secret"
	"github.com/go-arcade/arcade/internal/engine/repo/storage"
	"github.com/go-arcade/arcade/internal/engine/repo/team"
	"github.com/go-arcade/arcade/internal/engine/repo/user"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"gorm.io/gorm"
)

// Repositories 统一管理所有 repository
type Repositories struct {
	User                user.IUserRepository
	Agent               agent.IAgentRepository
	Plugin              plugin.IPluginRepository
	PluginTask          plugin.IPluginTaskRepository
	Permission          permission.IPermissionRepository
	Role                role.IRoleRepository
	Storage             storage.IStorageRepository
	Team                team.ITeamRepository
	IdentityIntegration identity_integration.IIdentityIntegrationRepository
	GeneralSettings     general_settings.IGeneralSettingsRepository
	RouterPermission    permission.IRouterPermissionRepository
	ProjectMember       project.IProjectMemberRepository
	ProjectTeamAccess   project.IProjectTeamAccessRepository
	TeamMember          team.ITeamMemberRepository
	UserExtension       user.IUserExtensionRepository
	Secret              secret.ISecretRepository
}

// NewRepositories 初始化所有 repository
func NewRepositories(db database.DB, mongo database.MongoDB, cache cache.Cache) *Repositories {
	return &Repositories{
		User:                user.NewUserRepo(db, cache),
		Agent:               agent.NewAgentRepo(db),
		Plugin:              plugin.NewPluginRepo(db),
		PluginTask:          plugin.NewPluginTaskRepo(mongo),
		Permission:          permission.NewPermissionRepo(db, cache),
		Role:                role.NewRoleRepo(db),
		Storage:             storage.NewStorageRepo(db, cache),
		Team:                team.NewTeamRepo(db),
		IdentityIntegration: identity_integration.NewIdentityIntegrationRepo(db),
		GeneralSettings:     general_settings.NewGeneralSettingsRepo(db),
		RouterPermission:    permission.NewRouterPermissionRepo(db),
		ProjectMember:       project.NewProjectMemberRepo(db),
		ProjectTeamAccess:   project.NewProjectTeamAccessRepo(db),
		TeamMember:          team.NewTeamMemberRepo(db),
		UserExtension:       user.NewUserExtensionRepo(db),
		Secret:              secret.NewSecretRepo(db),
	}
}

// GetDB 返回数据库实例（供插件适配器使用）
func (r *Repositories) GetDB() database.DB {
	return r.Storage.(*storage.StorageRepo).GetDB()
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
