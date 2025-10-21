package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/pkg/id"
	"github.com/observabil/arcade/pkg/log"
	"github.com/observabil/arcade/pkg/sso"
	"github.com/observabil/arcade/pkg/sso/ldap"
	"github.com/observabil/arcade/pkg/sso/oidc"
	"github.com/observabil/arcade/pkg/sso/util"
)

type AuthService struct {
	authRepo *repo.SSORepo
	userRepo *repo.UserRepo
}

func NewAuthService(authRepo *repo.SSORepo, userRepo *repo.UserRepo) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

func (as *AuthService) Redirect(providerName string) (string, error) {
	ssoProvider, err := as.authRepo.GetProvider(providerName)
	if err != nil {
		log.Errorf("failed to get oauth provider: %v", err)
		return "", err
	}

	var cfg sso.ProviderConfig
	if err := sonic.Unmarshal(ssoProvider.Config, &cfg); err != nil {
		return "", fmt.Errorf("unmarshal provider config failed: %w", err)
	}

	provider, err := sso.NewSSOProvider(&cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create SSO provider: %w", err)
	}

	state := util.GenState()
	util.StateStore.Store(state, providerName)

	switch ssoProvider.ProviderType {
	case "oauth":
		return provider.AuthCodeURL(state), nil
	case "oidc":
		return provider.AuthCodeURL(state), nil
	default:
		return "", fmt.Errorf("invalid provider type: %s", ssoProvider.ProviderType)
	}
}

// Callback 统一的 OAuth/OIDC 回调处理（根据 provider 类型自动识别）
func (as *AuthService) Callback(providerName, state, code string) (*model.Register, error) {
	storedProviderName, ok := util.LoadAndDeleteState(state)
	if !ok || storedProviderName != providerName {
		return nil, fmt.Errorf("invalid state parameter")
	}

	ssoProvider, err := as.authRepo.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("load provider failed: %w", err)
	}

	// 序列化
	var cfg sso.ProviderConfig
	if err := sonic.Unmarshal(ssoProvider.Config, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal provider config failed: %w", err)
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

	return as.registerOrLoginUser(providerName, &sso.UserInfoAdapter{
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		Nickname:  userInfo.Nickname,
		AvatarURL: userInfo.AvatarURL,
	})
}

func (as *AuthService) GetProviderByType(providerType string) ([]model.SSOProvider, error) {
	ssoProviders, err := as.authRepo.GetProviderByType(providerType)
	if err != nil {
		log.Errorf("failed to get provider by type: %v", err)
		return nil, err
	}
	return ssoProviders, nil
}

func (as *AuthService) GetProvider(name string) (*model.SSOProvider, error) {
	ssoProvider, err := as.authRepo.GetProvider(name)
	if err != nil {
		log.Errorf("failed to get provider: %v", err)
		return nil, err
	}
	return ssoProvider, nil
}

// GetProviderList 获取提供者列表
func (as *AuthService) GetProviderList() ([]model.SSOProvider, error) {
	ssoProviders, err := as.authRepo.GetProviderList()
	if err != nil {
		log.Errorf("failed to get provider list: %v", err)
		return nil, err
	}
	return ssoProviders, nil
}

func (as *AuthService) GetProviderTypeList() ([]string, error) {
	providerTypes, err := as.authRepo.GetProviderTypeList()
	if err != nil {
		log.Errorf("failed to get provider type list: %v", err)
		return nil, err
	}
	return providerTypes, nil
}

// OIDCLogin OIDC 登录
func (as *AuthService) OIDCLogin(providerName string) (string, error) {
	oidcConf, err := as.authRepo.GetProvider(providerName)
	if err != nil {
		log.Errorf("failed to get OIDC provider: %v", err)
		return "", err
	}

	var oidcCfg model.OIDCConfig
	if err := sonic.Unmarshal(oidcConf.Config, &oidcCfg); err != nil {
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
		log.Errorf("failed to create OIDC provider: %v", err)
		return "", err
	}

	state := util.GenState()
	util.StateStore.Store(state, providerName)

	return provider.GetAuthURL(state), nil
}

// LDAP 相关方法

// LDAPLogin LDAP 登录请求
type LDAPLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LDAPLogin LDAP 登录
func (as *AuthService) LDAPLogin(providerName, username, password string) (*model.Register, error) {
	ldapConf, err := as.authRepo.GetProvider(providerName)
	if err != nil {
		log.Errorf("failed to get LDAP provider: %v", err)
		return nil, err
	}

	var ldapCfg model.LDAPConfig
	if err := sonic.Unmarshal(ldapConf.Config, &ldapCfg); err != nil {
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
		log.Errorf("LDAP authentication failed: %v", err)
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// 使用通用的用户注册逻辑
	return as.registerOrLoginUser(providerName, &sso.UserInfoAdapter{
		Username: userInfo.Username,
		Email:    userInfo.Email,
		Nickname: userInfo.DisplayName,
	})
}

// registerOrLoginUser 通用的用户注册或登录逻辑
func (as *AuthService) registerOrLoginUser(providerName string, userInfo *sso.UserInfoAdapter) (*model.Register, error) {
	registerUserInfo := &model.Register{
		UserId:   id.GetUUIDWithoutDashes(),
		Username: userInfo.Username,
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.AvatarURL,
		Email:    userInfo.Email,
	}

	// 如果没有 email，生成默认 email
	if registerUserInfo.Email == "" {
		registerUserInfo.Email = fmt.Sprintf("%s@%s.com", userInfo.Username, providerName)
	}

	// 如果没有 nickname，使用 username
	if registerUserInfo.Nickname == "" {
		registerUserInfo.Nickname = userInfo.Username
	}

	registerUserInfo.CreateTime = time.Now()
	password, err := getPassword(userInfo.Username)
	if err != nil {
		return nil, err
	}
	registerUserInfo.Password = string(password)

	err = as.userRepo.Register(registerUserInfo)
	if err != nil {
		log.Errorf("failed to register user: %v", err)
		return nil, err
	}

	return registerUserInfo, nil
}
