package identity_integration

import (
	"github.com/go-arcade/arcade/internal/engine/model/identity_integration"
	"github.com/go-arcade/arcade/pkg/database"
)

type IIdentityIntegrationRepository interface {
	GetProvider(name string) (*identity_integration.IdentityIntegration, error)
	GetProviderByType(providerType string) ([]identity_integration.IdentityIntegration, error)
	GetProviderList() ([]identity_integration.IdentityIntegration, error)
	GetProviderTypeList() ([]string, error)
	CreateProvider(provider *identity_integration.IdentityIntegration) error
	UpdateProvider(name string, provider *identity_integration.IdentityIntegration) error
	DeleteProvider(name string) error
	ProviderExists(name string) (bool, error)
	ToggleProvider(name string) error
}

type IdentityIntegrationRepo struct {
	db          database.DB
	Integration identity_integration.IdentityIntegration
}

func NewIdentityIntegrationRepo(db database.DB) IIdentityIntegrationRepository {
	return &IdentityIntegrationRepo{
		db:          db,
		Integration: identity_integration.IdentityIntegration{},
	}
}

func (ii *IdentityIntegrationRepo) GetProvider(name string) (*identity_integration.IdentityIntegration, error) {
	var integration identity_integration.IdentityIntegration
	if err := ii.db.DB().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Select("provider_id, provider_type, name, description, config, priority, is_enabled").
		First(&integration).Error; err != nil {
		return nil, err
	}
	return &integration, nil
}

func (ii *IdentityIntegrationRepo) GetProviderByType(providerType string) ([]identity_integration.IdentityIntegration, error) {
	var integrations []identity_integration.IdentityIntegration
	if err := ii.db.DB().Table(ii.Integration.TableName()).
		Where("provider_type = ?", providerType).
		Order("priority ASC").
		Select("provider_id, provider_type, name, description, priority, is_enabled").
		Find(&integrations).Error; err != nil {
		return nil, err
	}
	return integrations, nil
}

func (ii *IdentityIntegrationRepo) GetProviderList() ([]identity_integration.IdentityIntegration, error) {
	var integrations []identity_integration.IdentityIntegration
	if err := ii.db.DB().Table(ii.Integration.TableName()).
		Order("priority ASC").
		Select("provider_id, provider_type, name, description, priority, is_enabled").
		Find(&integrations).Error; err != nil {
		return nil, err
	}
	return integrations, nil
}

func (ii *IdentityIntegrationRepo) GetProviderTypeList() ([]string, error) {
	var providerTypes []string
	if err := ii.db.DB().Table(ii.Integration.TableName()).
		Distinct("provider_type").
		Select("provider_type").
		Pluck("provider_type", &providerTypes).Error; err != nil {
		return nil, err
	}
	return providerTypes, nil
}

// CreateProvider creates an identity integration provider
func (ii *IdentityIntegrationRepo) CreateProvider(provider *identity_integration.IdentityIntegration) error {
	return ii.db.DB().Table(ii.Integration.TableName()).Create(provider).Error
}

// UpdateProvider updates an identity integration provider (name and provider_type fields cannot be updated)
func (ii *IdentityIntegrationRepo) UpdateProvider(name string, provider *identity_integration.IdentityIntegration) error {
	return ii.db.DB().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Omit("name", "provider_id", "provider_type", "created_at").
		Updates(provider).Error
}

// DeleteProvider deletes an identity integration provider
func (ii *IdentityIntegrationRepo) DeleteProvider(name string) error {
	return ii.db.DB().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Delete(&identity_integration.IdentityIntegration{}).Error
}

// ProviderExists checks if a provider exists
func (ii *IdentityIntegrationRepo) ProviderExists(name string) (bool, error) {
	var count int64
	err := ii.db.DB().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

// ToggleProvider toggles the enabled status of an identity integration provider (enable/disable)
func (ii *IdentityIntegrationRepo) ToggleProvider(name string) error {
	// query current status
	var integration identity_integration.IdentityIntegration
	if err := ii.db.DB().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Select("is_enabled").
		First(&integration).Error; err != nil {
		return err
	}

	// toggle status: 0 -> 1, 1 -> 0
	newStatus := 1 - integration.IsEnabled

	// update status
	return ii.db.DB().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Update("is_enabled", newStatus).Error
}
