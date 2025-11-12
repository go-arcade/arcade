package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IIdentityIntegrationRepository interface {
	GetProvider(name string) (*model.IdentityIntegration, error)
	GetProviderByType(providerType string) ([]model.IdentityIntegration, error)
	GetProviderList() ([]model.IdentityIntegration, error)
	GetProviderTypeList() ([]string, error)
	CreateProvider(provider *model.IdentityIntegration) error
	UpdateProvider(name string, provider *model.IdentityIntegration) error
	DeleteProvider(name string) error
	ProviderExists(name string) (bool, error)
	ToggleProvider(name string) error
}

type IdentityIntegrationRepo struct {
	db                       database.DB
	IdentityIntegrationModel model.IdentityIntegration
}

func NewIdentityIntegrationRepo(db database.DB) IIdentityIntegrationRepository {
	return &IdentityIntegrationRepo{
		db:                       db,
		IdentityIntegrationModel: model.IdentityIntegration{},
	}
}

func (ii *IdentityIntegrationRepo) GetProvider(name string) (*model.IdentityIntegration, error) {
	var integration model.IdentityIntegration
	if err := ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Where("name = ?", name).
		Select("provider_id, provider_type, name, description, config, priority, is_enabled").
		First(&integration).Error; err != nil {
		return nil, err
	}
	return &integration, nil
}

func (ii *IdentityIntegrationRepo) GetProviderByType(providerType string) ([]model.IdentityIntegration, error) {
	var integrations []model.IdentityIntegration
	if err := ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Where("provider_type = ?", providerType).
		Order("priority ASC").
		Select("provider_id, provider_type, name, description, priority, is_enabled").
		Find(&integrations).Error; err != nil {
		return nil, err
	}
	return integrations, nil
}

func (ii *IdentityIntegrationRepo) GetProviderList() ([]model.IdentityIntegration, error) {
	var integrations []model.IdentityIntegration
	if err := ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Order("priority ASC").
		Select("provider_id, provider_type, name, description, priority, is_enabled").
		Find(&integrations).Error; err != nil {
		return nil, err
	}
	return integrations, nil
}

func (ii *IdentityIntegrationRepo) GetProviderTypeList() ([]string, error) {
	var providerTypes []string
	if err := ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Distinct("provider_type").
		Select("provider_type").
		Pluck("provider_type", &providerTypes).Error; err != nil {
		return nil, err
	}
	return providerTypes, nil
}

// CreateProvider creates an identity integration provider
func (ii *IdentityIntegrationRepo) CreateProvider(provider *model.IdentityIntegration) error {
	return ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).Create(provider).Error
}

// UpdateProvider updates an identity integration provider (name and provider_type fields cannot be updated)
func (ii *IdentityIntegrationRepo) UpdateProvider(name string, provider *model.IdentityIntegration) error {
	return ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Where("name = ?", name).
		Omit("name", "provider_id", "provider_type", "created_at").
		Updates(provider).Error
}

// DeleteProvider deletes an identity integration provider
func (ii *IdentityIntegrationRepo) DeleteProvider(name string) error {
	return ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Where("name = ?", name).
		Delete(&model.IdentityIntegration{}).Error
}

// ProviderExists checks if a provider exists
func (ii *IdentityIntegrationRepo) ProviderExists(name string) (bool, error) {
	var count int64
	err := ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

// ToggleProvider toggles the enabled status of an identity integration provider (enable/disable)
func (ii *IdentityIntegrationRepo) ToggleProvider(name string) error {
	// query current status
	var integration model.IdentityIntegration
	if err := ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Where("name = ?", name).
		Select("is_enabled").
		First(&integration).Error; err != nil {
		return err
	}

	// toggle status: 0 -> 1, 1 -> 0
	newStatus := 1 - integration.IsEnabled

	// update status
	return ii.db.DB().Table(ii.IdentityIntegrationModel.TableName()).
		Where("name = ?", name).
		Update("is_enabled", newStatus).Error
}
