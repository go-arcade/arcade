package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IIdentityRepository interface {
	GetProvider(name string) (*model.Identity, error)
	GetProviderByType(providerType string) ([]model.Identity, error)
	GetProviderList() ([]model.Identity, error)
	GetProviderTypeList() ([]string, error)
	CreateProvider(provider *model.Identity) error
	UpdateProvider(name string, provider *model.Identity) error
	DeleteProvider(name string) error
	ProviderExists(name string) (bool, error)
	ToggleProvider(name string) error
}

type IdentityRepo struct {
	db          database.IDatabase
	Integration model.Identity
}

func NewIdentityRepo(db database.IDatabase) IIdentityRepository {
	return &IdentityRepo{
		db:          db,
		Integration: model.Identity{},
	}
}

func (ii *IdentityRepo) GetProvider(name string) (*model.Identity, error) {
	var integration model.Identity
	if err := ii.db.Database().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Select("provider_id, provider_type, name, description, config, priority, is_enabled").
		First(&integration).Error; err != nil {
		return nil, err
	}
	return &integration, nil
}

func (ii *IdentityRepo) GetProviderByType(providerType string) ([]model.Identity, error) {
	var integrations []model.Identity
	if err := ii.db.Database().Table(ii.Integration.TableName()).
		Where("provider_type = ?", providerType).
		Order("priority ASC").
		Select("provider_id, provider_type, name, description, priority, is_enabled").
		Find(&integrations).Error; err != nil {
		return nil, err
	}
	return integrations, nil
}

func (ii *IdentityRepo) GetProviderList() ([]model.Identity, error) {
	var integrations []model.Identity
	if err := ii.db.Database().Table(ii.Integration.TableName()).
		Order("priority ASC").
		Select("provider_id, provider_type, name, description, priority, is_enabled").
		Find(&integrations).Error; err != nil {
		return nil, err
	}
	return integrations, nil
}

func (ii *IdentityRepo) GetProviderTypeList() ([]string, error) {
	var providerTypes []string
	if err := ii.db.Database().Table(ii.Integration.TableName()).
		Distinct("provider_type").
		Select("provider_type").
		Pluck("provider_type", &providerTypes).Error; err != nil {
		return nil, err
	}
	return providerTypes, nil
}

// CreateProvider creates an identity provider
func (ii *IdentityRepo) CreateProvider(provider *model.Identity) error {
	return ii.db.Database().Table(ii.Integration.TableName()).Create(provider).Error
}

// UpdateProvider updates an identity provider (name and provider_type fields cannot be updated)
func (ii *IdentityRepo) UpdateProvider(name string, provider *model.Identity) error {
	return ii.db.Database().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Omit("name", "provider_id", "provider_type", "created_at").
		Updates(provider).Error
}

// DeleteProvider deletes an identity provider
func (ii *IdentityRepo) DeleteProvider(name string) error {
	return ii.db.Database().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Delete(&model.Identity{}).Error
}

// ProviderExists checks if a provider exists
func (ii *IdentityRepo) ProviderExists(name string) (bool, error) {
	var count int64
	err := ii.db.Database().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

// ToggleProvider toggles the enabled status of an identity provider (enable/disable)
func (ii *IdentityRepo) ToggleProvider(name string) error {
	// query current status
	var integration model.Identity
	if err := ii.db.Database().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Select("is_enabled").
		First(&integration).Error; err != nil {
		return err
	}

	// toggle status: 0 -> 1, 1 -> 0
	newStatus := 1 - integration.IsEnabled

	// update status
	return ii.db.Database().Table(ii.Integration.TableName()).
		Where("name = ?", name).
		Update("is_enabled", newStatus).Error
}
