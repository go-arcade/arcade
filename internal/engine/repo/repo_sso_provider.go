package repo

import (
	"encoding/json"
	"fmt"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
)

type SSORepo struct {
	Ctx             *ctx.Context
	AuthConfigModel model.SSOProvider
}

func NewSSORepo(ctx *ctx.Context) *SSORepo {
	return &SSORepo{
		Ctx:             ctx,
		AuthConfigModel: model.SSOProvider{},
	}
}

func (ar *SSORepo) GetOauthProvider(name string) (*model.OAuthConfig, error) {
	var ssoProvider model.SSOProvider
	if err := ar.Ctx.GetDB().Table(ar.AuthConfigModel.TableName()).
		Where("name = ? AND provider_type = ?", name, "oauth").
		First(&ssoProvider).Error; err != nil {
		return nil, err
	}

	var oauthConfig model.OAuthConfig
	if err := json.Unmarshal(ssoProvider.Config, &oauthConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oauth config: %v", err)
	}

	// 构建 oauth2.Endpoint
	if oauthConfig.Endpoint.AuthURL == "" {
		oauthConfig.Endpoint.AuthURL = oauthConfig.AuthURL
	}
	if oauthConfig.Endpoint.TokenURL == "" {
		oauthConfig.Endpoint.TokenURL = oauthConfig.TokenURL
	}

	return &oauthConfig, nil
}

func (ar *SSORepo) GetOauthProviderList() ([]model.SSOProvider, error) {
	var ssoProviders []model.SSOProvider
	if err := ar.Ctx.GetDB().Table(ar.AuthConfigModel.TableName()).
		Where("provider_type = ? AND is_enabled = ?", "oauth", 1).
		Order("priority ASC").
		Find(&ssoProviders).Error; err != nil {
		return nil, err
	}
	return ssoProviders, nil
}
