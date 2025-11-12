package service

import (
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/internal/engine/service/agent"
	generalsettingsservice "github.com/go-arcade/arcade/internal/engine/service/general_settings"
	identityservice "github.com/go-arcade/arcade/internal/engine/service/identity_integration"
	permissionservice "github.com/go-arcade/arcade/internal/engine/service/permission"
	serviceplugin "github.com/go-arcade/arcade/internal/engine/service/plugin"
	roleservice "github.com/go-arcade/arcade/internal/engine/service/role"
	secretservice "github.com/go-arcade/arcade/internal/engine/service/secret"
	storageservice "github.com/go-arcade/arcade/internal/engine/service/storage"
	teamservice "github.com/go-arcade/arcade/internal/engine/service/team"
	uploadservice "github.com/go-arcade/arcade/internal/engine/service/upload"
	userservice "github.com/go-arcade/arcade/internal/engine/service/user"
	storagepkg "github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
)

// Services 统一管理所有 service
type Services struct {
	User                *userservice.UserService
	Agent               *agent.AgentService
	IdentityIntegration *identityservice.IdentityIntegrationService
	Role                *roleservice.RoleService
	Team                *teamservice.TeamService
	Storage             *storageservice.StorageService
	Upload              *uploadservice.UploadService
	Secret              *secretservice.SecretService
	GeneralSettings     *generalsettingsservice.GeneralSettingsService
	UserExtension       *userservice.UserExtensionService
	Permission          *permissionservice.PermissionService
	UserPermissions     *userservice.UserPermissionsService
	Plugin              *serviceplugin.PluginService
}

// NewServices 初始化所有 service
func NewServices(
	ctx *ctx.Context,
	db database.DB,
	cache cache.Cache,
	repos *repo.Repositories,
	pluginManager *pluginpkg.Manager,
	storageProvider storagepkg.StorageProvider,
) *Services {
	// 基础服务
	userService := userservice.NewUserService(ctx, cache, repos.User)
	agentService := agent.NewAgentService(repos.Agent, nil)
	identityIntegrationService := identityservice.NewIdentityIntegrationService(repos.IdentityIntegration, repos.User)
	roleService := roleservice.NewRoleService(repos.Role)
	teamService := teamservice.NewTeamService(ctx, repos.Team)
	storageService := storageservice.NewStorageService(ctx, repos.Storage)
	uploadService := uploadservice.NewUploadService(ctx, repos.Storage)
	secretService := secretservice.NewSecretService(ctx, repos.Secret)
	generalSettingsService := generalsettingsservice.NewGeneralSettingsService(ctx, repos.GeneralSettings)
	userExtensionService := userservice.NewUserExtensionService(repos.UserExtension)
	permissionService := permissionservice.NewPermissionService(ctx, db, cache, repos.Permission, repos.Role, repos.User)

	// UserPermissionsService 依赖 PermissionService
	userPermissionsService := userservice.NewUserPermissionsService(ctx, db, permissionService, repos.RouterPermission)
	// PluginService 需要 pluginManager 和 storageProvider
	var pluginService *serviceplugin.PluginService
	if pluginManager != nil && storageProvider != nil {
		pluginService = serviceplugin.NewPluginService(ctx, repos.Plugin, pluginManager, storageProvider)
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
