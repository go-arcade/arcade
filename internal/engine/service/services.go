package service

import (
	"github.com/go-arcade/arcade/internal/engine/repo"
	storagepkg "github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"golang.org/x/crypto/bcrypt"
)

// Services 统一管理所有 service
type Services struct {
	User                *UserService
	Agent               *AgentService
	IdentityIntegration *IdentityIntegrationService
	Role                *RoleService
	Team                *TeamService
	Storage             *StorageService
	Upload              *UploadService
	Secret              *SecretService
	GeneralSettings     *GeneralSettingsService
	UserExtension       *UserExtensionService
	Permission          *PermissionService
	UserPermissions     *UserPermissionsService
	Plugin              *PluginService
}

// NewServices 初始化所有 service
func NewServices(
	ctx *ctx.Context,
	db database.IDatabase,
	cache cache.ICache,
	repos *repo.Repositories,
	pluginManager *pluginpkg.Manager,
	storageProvider storagepkg.StorageProvider,
) *Services {
	// 基础服务
	userService := NewUserService(ctx, cache, repos.User)
	agentService := NewAgentService(repos.Agent, nil)
	identityIntegrationService := NewIdentityIntegrationService(repos.IdentityIntegration, repos.User)
	roleService := NewRoleService(repos.Role)
	teamService := NewTeamService(ctx, repos.Team)
	storageService := NewStorageService(ctx, repos.Storage)
	uploadService := NewUploadService(ctx, repos.Storage)
	secretService := NewSecretService(ctx, repos.Secret)
	generalSettingsService := NewGeneralSettingsService(ctx, repos.GeneralSettings)
	userExtensionService := NewUserExtensionService(repos.UserExtension)
	permissionService := NewPermissionService(ctx, db, cache, repos.Permission, repos.Role, repos.User)

	// UserPermissionsService 依赖 PermissionService
	userPermissionsService := NewUserPermissionsService(ctx, db, permissionService, repos.RouterPermission)
	// PluginService 需要 pluginManager 和 storageProvider
	var pluginService *PluginService
	if pluginManager != nil && storageProvider != nil {
		pluginService = NewPluginService(ctx, repos.Plugin, pluginManager, storageProvider)
	}

	return &Services{
		User:                userService,
		Agent:               agentService,
		IdentityIntegration: identityIntegrationService,
		Role:                roleService,
		Team:                teamService,
		Storage:             storageService,
		Upload:              uploadService,
		Secret:              secretService,
		GeneralSettings:     generalSettingsService,
		UserExtension:       userExtensionService,
		Permission:          permissionService,
		UserPermissions:     userPermissionsService,
		Plugin:              pluginService,
	}
}

// getPassword generates a bcrypt hash for a password
func getPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, err
}
