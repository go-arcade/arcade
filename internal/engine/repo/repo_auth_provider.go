package repo

import (
	"encoding/json"
	"fmt"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/9 16:15
 * @file: repo_auth_provider.go
 * @description: repo auth provider
 */

type AuthRepo struct {
	Ctx             *ctx.Context
	AuthConfigModel model.OauthProvider
}

func NewAuthRepo(ctx *ctx.Context) *AuthRepo {
	return &AuthRepo{
		Ctx:             ctx,
		AuthConfigModel: model.OauthProvider{},
	}
}

func (ar *AuthRepo) GetOauthProvider(name string) (*model.OauthProviderContent, error) {
	var authConfig model.OauthProvider
	if err := ar.Ctx.GetDB().Table(ar.AuthConfigModel.TableName()).
		Select("id, name, content").Where("name = ?", name).
		First(&authConfig).Error; err != nil {
		return nil, err
	}

	var content model.OauthProviderContent
	if err := json.Unmarshal(authConfig.Content, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %v", err)
	}
	return &content, nil
}

func (ar *AuthRepo) GetOauthProviderList() ([]model.OauthProvider, error) {
	var authConfigs []model.OauthProvider
	var err error
	if err = ar.Ctx.GetDB().Table(ar.AuthConfigModel.TableName()).
		Select("id, name, content").Find(&authConfigs).
		Error; err != nil {
		return nil, err
	}
	return authConfigs, nil
}
