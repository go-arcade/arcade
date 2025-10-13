package service

import (
	"context"
	"fmt"
	"time"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/pkg/id"
	"github.com/observabil/arcade/pkg/log"
	"github.com/observabil/arcade/pkg/sso/oauth"
	"golang.org/x/oauth2"
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

func (as *AuthService) Oauth(providerName string) (string, error) {

	oauthConf, err := as.authRepo.GetOauthProvider(providerName)
	if err != nil {
		log.Errorf("failed to get oauth provider: %v", err)
		return "", err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     oauthConf.ClientID,
		ClientSecret: oauthConf.ClientSecret,
		RedirectURL:  oauthConf.RedirectURL,
		Scopes:       oauthConf.Scopes,
		Endpoint:     oauthConf.Endpoint,
	}

	if oauthConfig.Endpoint.AuthURL == "" || oauthConfig.Endpoint.TokenURL == "" {
		return "", fmt.Errorf("invalid OAuth endpoints for provider: %s", providerName)
	}

	state := oauth.GenState()
	oauth.StateStore.Store(state, providerName)

	return oauthConfig.AuthCodeURL(state), nil
}

func (as *AuthService) Callback(providerName, state, code string) (*model.Register, error) {

	storedProviderName, ok := oauth.LoadAndDeleteState(state)
	if !ok || storedProviderName != providerName {
		return nil, fmt.Errorf("invalid state parameter")
	}

	oauthConf, err := as.authRepo.GetOauthProvider(providerName)
	if err != nil {
		log.Errorf("failed to get oauth provider: %v", err)
		return nil, err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     oauthConf.ClientID,
		ClientSecret: oauthConf.ClientSecret,
		RedirectURL:  oauthConf.RedirectURL,
		Scopes:       oauthConf.Scopes,
		Endpoint:     oauthConf.Endpoint,
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Errorf("failed to exchange token: %v", err)
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	getUserInfoFunc, ok := oauth.GetUserInfoFunc[oauth.Provider(providerName)]
	if !ok {
		log.Errorf("unsupported provider: %s", providerName)
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	userInfo, err := getUserInfoFunc(token)
	if err != nil {
		log.Errorf("failed to get user information: %v", err)
		return nil, fmt.Errorf("failed to obtain user information: %w", err)
	}

	registerUserInfo := &model.Register{
		UserId:   id.GetUUIDWithoutDashes(),
		Username: userInfo.Username,
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.AvatarURL,
	}
	registerUserInfo.Email = fmt.Sprintf("%s@%s.com", userInfo.Username, providerName)
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

func (as *AuthService) GetOauthProvider(name string) (*model.OAuthConfig, error) {

	oauthConfig, err := as.authRepo.GetOauthProvider(name)
	if err != nil {
		log.Errorf("failed to get oauth provider: %v", err)
		return nil, err
	}
	return oauthConfig, nil
}

func (as *AuthService) GetOauthProviderList() ([]model.SSOProvider, error) {

	authConfigs, err := as.authRepo.GetOauthProviderList()
	return authConfigs, err
}
