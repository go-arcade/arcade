package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/database"
)

type IGeneralSettingsRepository interface {
	UpdateGeneralSettings(settings *model.GeneralSettings) error
	GetGeneralSettingsByID(id uint64) (*model.GeneralSettings, error)
	GetGeneralSettingsByName(category, name string) (*model.GeneralSettings, error)
	GetGeneralSettingsList(pageNum, pageSize int, category string) ([]*model.GeneralSettings, int64, error)
	GetCategories() ([]string, error)
}

type GeneralSettingsRepo struct {
	database.IDatabase
}

func NewGeneralSettingsRepo(db database.IDatabase) IGeneralSettingsRepository {
	return &GeneralSettingsRepo{
		IDatabase: db,
	}
}

// UpdateGeneralSettings updates a general settings by ID
func (gsr *GeneralSettingsRepo) UpdateGeneralSettings(settings *model.GeneralSettings) error {
	return gsr.Database().Table(settings.TableName()).
		Omit("id", "category", "name").
		Where("id = ?", settings.ID).
		Updates(settings).Error
}

// GetGeneralSettingsByID gets a general settings by ID
func (gsr *GeneralSettingsRepo) GetGeneralSettingsByID(id uint64) (*model.GeneralSettings, error) {
	var settings model.GeneralSettings
	err := gsr.Database().Table(settings.TableName()).
		Where("id = ?", id).
		First(&settings).Error
	return &settings, err
}

// GetGeneralSettingsByName gets a general settings by category and name
func (gsr *GeneralSettingsRepo) GetGeneralSettingsByName(category, name string) (*model.GeneralSettings, error) {
	var settings model.GeneralSettings
	err := gsr.Database().Table(settings.TableName()).
		Where("category = ? AND name = ?", category, name).
		First(&settings).Error
	return &settings, err
}

// GetGeneralSettingsList gets general settings list with pagination and filters
func (gsr *GeneralSettingsRepo) GetGeneralSettingsList(pageNum, pageSize int, category string) ([]*model.GeneralSettings, int64, error) {
	var settingsList []*model.GeneralSettings
	var settings model.GeneralSettings
	var total int64

	query := gsr.Database().Table(settings.TableName())

	// apply filters
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// get paginated list (specify fields, exclude create_time and update_time)
	offset := (pageNum - 1) * pageSize
	err := query.Select("id", "category", "name", "display_name", "data", "schema", "description").
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&settingsList).Error

	return settingsList, total, err
}

// GetCategories gets all distinct categories
func (gsr *GeneralSettingsRepo) GetCategories() ([]string, error) {
	var categories []string
	var settings model.GeneralSettings
	err := gsr.Database().Table(settings.TableName()).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}
