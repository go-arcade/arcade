package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	identitymodel "github.com/go-arcade/arcade/internal/engine/model"
	userrepo "github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/sso"
	"github.com/go-arcade/arcade/pkg/sso/ldap"
	"github.com/go-arcade/arcade/pkg/sso/oidc"
	"github.com/go-arcade/arcade/pkg/sso/util"
)

type IdentityIntegrationService struct {
	identityRepo userrepo.IIdentityIntegrationRepository
	userRepo     userrepo.IUserRepository
}

func NewIdentityIntegrationService(identityRepo userrepo.IIdentityIntegrationRepository, userRepo userrepo.IUserRepository) *IdentityIntegrationService {
	return &IdentityIntegrationService{
		identityRepo: identityRepo,
		userRepo:     userRepo,
	}
}

func (iis *IdentityIntegrationService) Authorize(providerName string) (string, error) {
	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		log.Error("failed to get oauth provider: %v", err)
		return "", err
	}

	var cfg sso.ProviderConfig
	if err := sonic.Unmarshal(integration.Config, &cfg); err != nil {
		return "", fmt.Errorf("unmarshal provider config failed: %w", err)
	}

	provider, err := sso.NewSSOProvider(&cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create SSO provider: %w", err)
	}

	state := util.GenState()
	util.StateStore.Store(state, providerName)

	switch integration.ProviderType {
	case "oauth":
		return provider.AuthCodeURL(state), nil
	case "oidc":
		return provider.AuthCodeURL(state), nil
	default:
		return "", fmt.Errorf("invalid provider type: %s", integration.ProviderType)
	}
}

// Callback 统一的 OAuth/OIDC 回调处理（根据 provider 类型自动识别）
func (iis *IdentityIntegrationService) Callback(providerName, state, code string) (*identitymodel.Register, error) {
	storedProviderName, ok := util.LoadAndDeleteState(state)
	if !ok || storedProviderName != providerName {
		log.Warn("invalid state parameter for %s, storedProviderName: %s, providerName: %s, state: %s",
			providerName, storedProviderName, providerName, state)
		return nil, fmt.Errorf("invalid state parameter")
	}

	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("load provider failed: %w", err)
	}

	// 序列化
	var cfg sso.ProviderConfig
	if err := sonic.Unmarshal(integration.Config, &cfg); err != nil {
		log.Error("invalid provider config for %s: %v", providerName, err)
		return nil, fmt.Errorf("provider config invalid")
	}

	provider, err := sso.NewSSOProvider(&cfg)
	if err != nil {
		return nil, fmt.Errorf("create provider failed: %w", err)
	}

	token, err := provider.ExchangeToken(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	userInfo, err := provider.GetUserInfo(context.Background(), token)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}

	return iis.registerOrLoginUser(providerName, &sso.UserInfoAdapter{
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		Nickname:  userInfo.Nickname,
		AvatarURL: userInfo.AvatarURL,
	})
}

// splitName splits a full name into first and last name
func splitName(name, nickname string) (firstName, lastName string) {
	// use name first, fallback to nickname
	fullName := name
	if fullName == "" {
		fullName = nickname
	}

	if fullName == "" {
		return "", ""
	}

	// split by space
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	}

	firstName = parts[0]
	if len(parts) > 1 {
		lastName = strings.Join(parts[1:], " ")
	}

	return firstName, lastName
}

func (iis *IdentityIntegrationService) GetProviderByType(providerType string) ([]identitymodel.IdentityIntegration, error) {
	integrations, err := iis.identityRepo.GetProviderByType(providerType)
	if err != nil {
		log.Error("failed to get provider by type: %v", err)
		return nil, err
	}
	return integrations, nil
}

func (iis *IdentityIntegrationService) GetProvider(name string) (*identitymodel.IdentityIntegration, error) {
	integration, err := iis.identityRepo.GetProvider(name)
	if err != nil {
		log.Error("failed to get provider: %v", err)
		return nil, err
	}
	return integration, nil
}

// GetProviderList 获取提供者列表
func (iis *IdentityIntegrationService) GetProviderList() ([]identitymodel.IdentityIntegration, error) {
	integrations, err := iis.identityRepo.GetProviderList()
	if err != nil {
		log.Error("failed to get provider list: %v", err)
		return nil, err
	}
	return integrations, nil
}

func (iis *IdentityIntegrationService) GetProviderTypeList() ([]string, error) {
	providerTypes, err := iis.identityRepo.GetProviderTypeList()
	if err != nil {
		log.Error("failed to get provider type list: %v", err)
		return nil, err
	}
	return providerTypes, nil
}

// OIDCLogin OIDC 登录
func (iis *IdentityIntegrationService) OIDCLogin(providerName string) (string, error) {
	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		log.Error("failed to get OIDC provider: %v", err)
		return "", err
	}

	var oidcCfg identitymodel.OIDCConfig
	if err := sonic.Unmarshal(integration.Config, &oidcCfg); err != nil {
		return "", fmt.Errorf("unmarshal provider config failed: %w", err)
	}

	// 创建 OIDC provider
	provider, err := oidc.NewOIDCProvider(
		oidcCfg.Issuer,
		oidcCfg.ClientID,
		oidcCfg.ClientSecret,
		oidcCfg.RedirectURL,
		oidcCfg.Scopes,
		oidcCfg.SkipVerify,
	)
	if err != nil {
		log.Error("failed to create OIDC provider: %v", err)
		return "", err
	}

	state := util.GenState()
	util.StateStore.Store(state, providerName)

	return provider.GetAuthURL(state), nil
}

// LDAP 相关方法

type LDAPLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LDAPLogin LDAP 登录
func (iis *IdentityIntegrationService) LDAPLogin(providerName, username, password string) (*identitymodel.Register, error) {
	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		log.Error("failed to get LDAP provider: %v", err)
		return nil, err
	}

	var ldapCfg identitymodel.LDAPConfig
	if err := sonic.Unmarshal(integration.Config, &ldapCfg); err != nil {
		return nil, fmt.Errorf("unmarshal provider config failed: %w", err)
	}

	// 创建 LDAP 客户端
	ldapClient := ldap.NewLDAPClient(
		ldapCfg.Host,
		ldapCfg.Port,
		ldapCfg.UseTLS,
		ldapCfg.SkipVerify,
		ldapCfg.BaseDN,
		ldapCfg.BindDN,
		ldapCfg.BindPassword,
	)

	// 认证用户
	userInfo, err := ldapClient.Authenticate(username, password)
	if err != nil {
		log.Error("LDAP authentication failed: %v", err)
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// use common user registration logic
	return iis.registerOrLoginUser(providerName, &sso.UserInfoAdapter{
		Username: userInfo.Username,
		Email:    userInfo.Email,
		Name:     userInfo.DisplayName,
		Nickname: userInfo.DisplayName,
	})
}

// registerOrLoginUser common user registration or login logic
func (iis *IdentityIntegrationService) registerOrLoginUser(providerName string, userInfo *sso.UserInfoAdapter) (*identitymodel.Register, error) {
	// split name into first and last name
	firstName, lastName := splitName(userInfo.Name, userInfo.Nickname)

	registerUserInfo := &identitymodel.Register{
		UserId:    id.GetUUIDWithoutDashes(),
		Username:  userInfo.Username,
		FirstName: firstName,
		LastName:  lastName,
		Avatar:    userInfo.AvatarURL,
		Email:     userInfo.Email,
	}

	// if no email, generate default email
	if registerUserInfo.Email == "" {
		registerUserInfo.Email = fmt.Sprintf("%s@%s.com", userInfo.Username, providerName)
	}

	// if no first name, use username as first name
	if registerUserInfo.FirstName == "" {
		registerUserInfo.FirstName = userInfo.Username
	}

	registerUserInfo.CreateTime = time.Now()
	password, err := getPassword(userInfo.Username)
	if err != nil {
		return nil, err
	}
	registerUserInfo.Password = string(password)

	err = iis.userRepo.Register(registerUserInfo)
	if err != nil {
		log.Error("failed to register user: %v", err)
		return nil, err
	}

	return registerUserInfo, nil
}

// CreateProvider creates an identity integration provider
func (iis *IdentityIntegrationService) CreateProvider(provider *identitymodel.IdentityIntegration) error {
	// check if provider name already exists
	exists, err := iis.identityRepo.ProviderExists(provider.Name)
	if err != nil {
		log.Error("failed to check provider exists: %v", err)
		return err
	}
	if exists {
		return fmt.Errorf("provider name already exists: %s", provider.Name)
	}

	// generate provider ID
	provider.ProviderId = id.GetUUID()

	if err := iis.identityRepo.CreateProvider(provider); err != nil {
		log.Error("failed to create provider: %v", err)
		return err
	}

	return nil
}

// UpdateProvider updates an identity integration provider
func (iis *IdentityIntegrationService) UpdateProvider(name string, provider *identitymodel.IdentityIntegration) error {
	// check if provider exists
	existing, err := iis.identityRepo.GetProvider(name)
	if err != nil {
		log.Error("failed to get provider: %v", err)
		return fmt.Errorf("provider not found: %s", name)
	}

	// preserve immutable configuration fields
	if err := iis.preserveConfigKeys(existing, provider); err != nil {
		return err
	}

	if err := iis.identityRepo.UpdateProvider(name, provider); err != nil {
		log.Error("failed to update provider: %v", err)
		return err
	}

	return nil
}

// preserveConfigKeys ensures key fields in config cannot be updated
func (iis *IdentityIntegrationService) preserveConfigKeys(existing, updated *identitymodel.IdentityIntegration) error {
	// unmarshal existing config
	var existingConfig map[string]interface{}
	if err := sonic.Unmarshal(existing.Config, &existingConfig); err != nil {
		return fmt.Errorf("failed to unmarshal existing config: %w", err)
	}

	// unmarshal updated config
	var updatedConfig map[string]interface{}
	if err := sonic.Unmarshal(updated.Config, &updatedConfig); err != nil {
		return fmt.Errorf("failed to unmarshal updated config: %w", err)
	}

	// define immutable keys based on provider type
	immutableKeys := getImmutableConfigKeys(existing.ProviderType)

	// preserve immutable keys from existing config
	for _, key := range immutableKeys {
		if existingValue, exists := existingConfig[key]; exists {
			updatedConfig[key] = existingValue
		}
	}

	// marshal back to JSON
	newConfig, err := sonic.Marshal(updatedConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	updated.Config = newConfig
	return nil
}

// getImmutableConfigKeys returns list of immutable config keys for each provider type
func getImmutableConfigKeys(providerType string) []string {
	switch providerType {
	case "oauth":
		return []string{"clientId"}
	case "oidc":
		return []string{"clientId", "issuer"}
	case "ldap":
		return []string{"host", "port", "baseDN"}
	default:
		return []string{}
	}
}

// DeleteProvider deletes an identity integration provider
func (iis *IdentityIntegrationService) DeleteProvider(name string) error {
	// check if provider exists
	exists, err := iis.identityRepo.ProviderExists(name)
	if err != nil {
		log.Error("failed to check provider exists: %v", err)
		return err
	}
	if !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	if err := iis.identityRepo.DeleteProvider(name); err != nil {
		log.Error("failed to delete provider: %v", err)
		return err
	}

	return nil
}

// ToggleProvider toggles the enabled status of an identity integration provider
func (iis *IdentityIntegrationService) ToggleProvider(name string) error {
	// check if provider exists
	exists, err := iis.identityRepo.ProviderExists(name)
	if err != nil {
		log.Error("failed to check provider exists: %v", err)
		return err
	}
	if !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	if err := iis.identityRepo.ToggleProvider(name); err != nil {
		log.Error("failed to toggle provider: %v", err)
		return err
	}

	return nil
}
