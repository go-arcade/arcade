// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"golang.org/x/oauth2"
)

type IdentityService struct {
	identityRepo userrepo.IIdentityRepository
	userRepo     userrepo.IUserRepository
}

func NewIdentityService(identityRepo userrepo.IIdentityRepository, userRepo userrepo.IUserRepository) *IdentityService {
	return &IdentityService{
		identityRepo: identityRepo,
		userRepo:     userRepo,
	}
}

func (iis *IdentityService) Authorize(providerName string) (string, error) {
	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		log.Errorw("failed to get oauth provider", "provider", providerName, "error", err)
		return "", err
	}

	cfg, err := iis.convertToProviderConfig(integration)
	if err != nil {
		return "", fmt.Errorf("convert provider config failed: %w", err)
	}

	provider, err := sso.NewSSOProvider(cfg)
	if err != nil {
		log.Errorw("failed to create SSO provider", "provider", providerName, "error", err)
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
func (iis *IdentityService) Callback(providerName, state, code string) (*identitymodel.Register, error) {
	storedProviderName, ok := util.LoadAndDeleteState(state)
	if !ok || storedProviderName != providerName {
		log.Warnw("invalid state parameter", "provider", providerName, "storedProviderName", storedProviderName, "state", state)
		return nil, fmt.Errorf("invalid state parameter")
	}

	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("load provider failed: %w", err)
	}

	cfg, err := iis.convertToProviderConfig(integration)
	if err != nil {
		log.Errorw("invalid provider config", "provider", providerName, "error", err)
		return nil, fmt.Errorf("provider config invalid: %w", err)
	}

	provider, err := sso.NewSSOProvider(cfg)
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
func splitName(name, nickname string) string {
	// use name first, fallback to nickname
	fullName := name
	if fullName == "" {
		fullName = nickname
	}

	if fullName == "" {
		return ""
	}

	// split by space
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

// convertToProviderConfig converts database config to sso.ProviderConfig
func (iis *IdentityService) convertToProviderConfig(integration *identitymodel.Identity) (*sso.ProviderConfig, error) {
	cfg := &sso.ProviderConfig{
		Type: integration.ProviderType,
	}

	switch integration.ProviderType {
	case "oauth":
		var oauthCfg identitymodel.OAuthConfig
		if err := sonic.Unmarshal(integration.Config, &oauthCfg); err != nil {
			return nil, fmt.Errorf("unmarshal oauth config failed: %w", err)
		}

		cfg.ClientID = oauthCfg.ClientID
		cfg.ClientSecret = oauthCfg.ClientSecret
		cfg.RedirectURL = oauthCfg.RedirectURL
		cfg.Scopes = oauthCfg.Scopes
		cfg.UserInfoURL = oauthCfg.UserInfoURL

		// Convert AuthURL and TokenURL to Endpoint
		if oauthCfg.Endpoint.AuthURL != "" && oauthCfg.Endpoint.TokenURL != "" {
			cfg.Endpoint = oauthCfg.Endpoint
		} else if oauthCfg.AuthURL != "" && oauthCfg.TokenURL != "" {
			cfg.Endpoint = oauth2.Endpoint{
				AuthURL:  oauthCfg.AuthURL,
				TokenURL: oauthCfg.TokenURL,
			}
		} else {
			return nil, fmt.Errorf("missing authURL or tokenURL in oauth config")
		}

	case "oidc":
		var oidcCfg identitymodel.OIDCConfig
		if err := sonic.Unmarshal(integration.Config, &oidcCfg); err != nil {
			return nil, fmt.Errorf("unmarshal oidc config failed: %w", err)
		}

		cfg.Issuer = oidcCfg.Issuer
		cfg.ClientID = oidcCfg.ClientID
		cfg.ClientSecret = oidcCfg.ClientSecret
		cfg.RedirectURL = oidcCfg.RedirectURL
		cfg.Scopes = oidcCfg.Scopes
		cfg.SkipVerify = oidcCfg.SkipVerify

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", integration.ProviderType)
	}

	return cfg, nil
}

func (iis *IdentityService) GetProviderByType(providerType string) ([]identitymodel.Identity, error) {
	integrations, err := iis.identityRepo.GetProviderByType(providerType)
	if err != nil {
		log.Errorw("failed to get provider by type", "providerType", providerType, "error", err)
		return nil, err
	}
	return integrations, nil
}

func (iis *IdentityService) GetProvider(name string) (*identitymodel.Identity, error) {
	integration, err := iis.identityRepo.GetProvider(name)
	if err != nil {
		log.Errorw("failed to get provider", "name", name, "error", err)
		return nil, err
	}
	return integration, nil
}

// GetProviderList 获取提供者列表
func (iis *IdentityService) GetProviderList() ([]identitymodel.Identity, error) {
	integrations, err := iis.identityRepo.GetProviderList()
	if err != nil {
		log.Errorw("failed to get provider list", "error", err)
		return nil, err
	}
	return integrations, nil
}

func (iis *IdentityService) GetProviderTypeList() ([]string, error) {
	providerTypes, err := iis.identityRepo.GetProviderTypeList()
	if err != nil {
		log.Errorw("failed to get provider type list", "error", err)
		return nil, err
	}
	return providerTypes, nil
}

// OIDCLogin OIDC 登录
func (iis *IdentityService) OIDCLogin(providerName string) (string, error) {
	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		log.Errorw("failed to get OIDC provider", "provider", providerName, "error", err)
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
		log.Errorw("failed to create OIDC provider", "provider", providerName, "error", err)
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
func (iis *IdentityService) LDAPLogin(providerName, username, password string) (*identitymodel.Register, error) {
	integration, err := iis.identityRepo.GetProvider(providerName)
	if err != nil {
		log.Errorw("failed to get LDAP provider", "provider", providerName, "error", err)
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
		log.Errorw("LDAP authentication failed", "provider", providerName, "username", username, "error", err)
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
func (iis *IdentityService) registerOrLoginUser(providerName string, userInfo *sso.UserInfoAdapter) (*identitymodel.Register, error) {
	// split name into first and last name
	fullName := splitName(userInfo.Name, userInfo.Nickname)

	registerUserInfo := &identitymodel.Register{
		UserId:   id.GetUUIDWithoutDashes(),
		Username: userInfo.Username,
		FullName: fullName,
		Avatar:   userInfo.AvatarURL,
		Email:    userInfo.Email,
	}

	// if no email, generate default email
	if registerUserInfo.Email == "" {
		registerUserInfo.Email = fmt.Sprintf("%s@%s.com", userInfo.Username, providerName)
	}

	// if no first name, use username as first name
	if registerUserInfo.FullName == "" {
		registerUserInfo.FullName = userInfo.Username
	}

	registerUserInfo.CreateTime = time.Now()
	password, err := getPassword(userInfo.Username)
	if err != nil {
		return nil, err
	}
	registerUserInfo.Password = string(password)

	err = iis.userRepo.Register(registerUserInfo)
	if err != nil {
		log.Errorw("failed to register user", "provider", providerName, "username", userInfo.Username, "error", err)
		return nil, err
	}

	return registerUserInfo, nil
}

// CreateProvider creates an identity provider
func (iis *IdentityService) CreateProvider(provider *identitymodel.Identity) error {
	// check if provider name already exists
	exists, err := iis.identityRepo.ProviderExists(provider.Name)
	if err != nil {
		log.Errorw("failed to check provider exists", "name", provider.Name, "error", err)
		return err
	}
	if exists {
		return fmt.Errorf("provider name already exists: %s", provider.Name)
	}

	// generate provider ID
	provider.ProviderId = id.GetUUID()

	if err := iis.identityRepo.CreateProvider(provider); err != nil {
		log.Errorw("failed to create provider", "name", provider.Name, "error", err)
		return err
	}

	return nil
}

// UpdateProvider updates an identity provider
func (iis *IdentityService) UpdateProvider(name string, provider *identitymodel.Identity) error {
	// check if provider exists
	existing, err := iis.identityRepo.GetProvider(name)
	if err != nil {
		log.Errorw("failed to get provider", "name", name, "error", err)
		return fmt.Errorf("provider not found: %s", name)
	}

	// preserve immutable configuration fields
	if err := iis.preserveConfigKeys(existing, provider); err != nil {
		return err
	}

	if err := iis.identityRepo.UpdateProvider(name, provider); err != nil {
		log.Errorw("failed to update provider", "name", name, "error", err)
		return err
	}

	return nil
}

// preserveConfigKeys ensures key fields in config cannot be updated
func (iis *IdentityService) preserveConfigKeys(existing, updated *identitymodel.Identity) error {
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

// DeleteProvider deletes an identity provider
func (iis *IdentityService) DeleteProvider(name string) error {
	// check if provider exists
	exists, err := iis.identityRepo.ProviderExists(name)
	if err != nil {
		log.Errorw("failed to check provider exists", "name", name, "error", err)
		return err
	}
	if !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	if err := iis.identityRepo.DeleteProvider(name); err != nil {
		log.Errorw("failed to delete provider", "name", name, "error", err)
		return err
	}

	return nil
}

// ToggleProvider toggles the enabled status of an identity provider
func (iis *IdentityService) ToggleProvider(name string) error {
	// check if provider exists
	exists, err := iis.identityRepo.ProviderExists(name)
	if err != nil {
		log.Errorw("failed to check provider exists", "name", name, "error", err)
		return err
	}
	if !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	if err := iis.identityRepo.ToggleProvider(name); err != nil {
		log.Errorw("failed to toggle provider", "name", name, "error", err)
		return err
	}

	return nil
}
