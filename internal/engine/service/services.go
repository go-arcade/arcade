package service

import (
	"github.com/go-arcade/arcade/internal/engine/repo"
	storagepkg "github.com/go-arcade/arcade/internal/pkg/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	pluginpkg "github.com/go-arcade/arcade/pkg/plugin"
	"golang.org/x/crypto/bcrypt"
)

// Services 统一管理所有 service
type Services struct {
	User            *UserService
	Agent           *AgentService
	Identity        *IdentityService
	Role            *RoleService
	Team            *TeamService
	Storage         *StorageService
	Upload          *UploadService
	Secret          *SecretService
	GeneralSettings *GeneralSettingsService
	UserExtension   *UserExtensionService
	Permission      *PermissionService
	UserPermissions *UserPermissionsService
	Plugin          *PluginService
}

// NewServices 初始化所有 service
func NewServices(
	db database.IDatabase,
	cache cache.ICache,
	repos *repo.Repositories,
	pluginManager *pluginpkg.Manager,
	storageProvider storagepkg.StorageProvider,
) *Services {
	// 基础服务
	userService := NewUserService(cache, repos.User)
	generalSettingsService := NewGeneralSettingsService(repos.GeneralSettings)
	agentService := NewAgentService(repos.Agent, generalSettingsService)
	identityService := NewIdentityService(repos.Identity, repos.User)
	roleService := NewRoleService(repos.Role)
	teamService := NewTeamService(repos.Team)
	storageService := NewStorageService(repos.Storage)
	uploadService := NewUploadService(repos.Storage)
	secretService := NewSecretService(repos.Secret)
	userExtensionService := NewUserExtensionService(repos.UserExtension)
	permissionService := NewPermissionService(db, cache, repos.Permission, repos.Role, repos.User)

	// UserPermissionsService 依赖 PermissionService
	userPermissionsService := NewUserPermissionsService(db, permissionService, repos.RouterPermission)
	// PluginService 需要 pluginManager 和 storageProvider
	var pluginService *PluginService
	if pluginManager != nil && storageProvider != nil {
		pluginService = NewPluginService(repos.Plugin, pluginManager, storageProvider)
	}

	return &Services{
		User:            userService,
		Agent:           agentService,
		Identity:        identityService,
		Role:            roleService,
		Team:            teamService,
		Storage:         storageService,
		Upload:          uploadService,
		Secret:          secretService,
		GeneralSettings: generalSettingsService,
		UserExtension:   userExtensionService,
		Permission:      permissionService,
		UserPermissions: userPermissionsService,
		Plugin:          pluginService,
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
