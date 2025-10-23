package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/ctx"
)

type SSORepo struct {
	Ctx              *ctx.Context
	SSOProviderModel model.SSOProvider
}

func NewSSORepo(ctx *ctx.Context) *SSORepo {
	return &SSORepo{
		Ctx:              ctx,
		SSOProviderModel: model.SSOProvider{},
	}
}

func (ar *SSORepo) GetProvider(name string) (*model.SSOProvider, error) {
	var ssoProvider model.SSOProvider
	if err := ar.Ctx.DBSession().Table(ar.SSOProviderModel.TableName()).
		Where("name = ?", name).
		Select("provider_type, name, description, priority, is_enabled").
		First(&ssoProvider).Error; err != nil {
		return nil, err
	}
	return &ssoProvider, nil
}

func (ar *SSORepo) GetProviderByType(providerType string) ([]model.SSOProvider, error) {
	var ssoProviders []model.SSOProvider
	if err := ar.Ctx.DBSession().Table(ar.SSOProviderModel.TableName()).
		Where("provider_type = ? AND is_enabled = ?", providerType, 1).
		Order("priority ASC").
		Select("provider_type, name, description, priority, is_enabled").
		Find(&ssoProviders).Error; err != nil {
		return nil, err
	}
	return ssoProviders, nil
}

func (ar *SSORepo) GetProviderList() ([]model.SSOProvider, error) {
	var ssoProviders []model.SSOProvider
	if err := ar.Ctx.DBSession().Table(ar.SSOProviderModel.TableName()).
		Where("is_enabled = ?", 1).
		Order("priority ASC").
		Select("provider_type, name, description, priority, is_enabled").
		Find(&ssoProviders).Error; err != nil {
		return nil, err
	}
	return ssoProviders, nil
}

func (ar *SSORepo) GetProviderTypeList() ([]string, error) {
	var providerTypes []string
	if err := ar.Ctx.DBSession().Table(ar.SSOProviderModel.TableName()).
		Distinct("provider_type").
		Where("is_enabled = ?", 1).
		Select("provider_type").
		Pluck("provider_type", &providerTypes).Error; err != nil {
		return nil, err
	}
	return providerTypes, nil
}
