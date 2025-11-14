package general_settings

import (
	"github.com/go-arcade/arcade/internal/engine/model/general_settings"
	"github.com/go-arcade/arcade/pkg/database"
)


type IGeneralSettingsRepository interface {
	UpdateGeneralSettings(settings *general_settings.GeneralSettings) error
	GetGeneralSettingsByID(id uint64) (*general_settings.GeneralSettings, error)
	GetGeneralSettingsByName(category, name string) (*general_settings.GeneralSettings, error)
	GetGeneralSettingsList(pageNum, pageSize int, category string) ([]*general_settings.GeneralSettings, int64, error)
	GetCategories() ([]string, error)
}

type GeneralSettingsRepo struct {
	db                   database.DB
	generalSettingsModel *general_settings.GeneralSettings
}

func NewGeneralSettingsRepo(db database.DB) IGeneralSettingsRepository {
	return &GeneralSettingsRepo{
		db:                   db,
		generalSettingsModel: &general_settings.GeneralSettings{},
	}
}

// UpdateGeneralSettings updates a general settings by ID
func (gsr *GeneralSettingsRepo) UpdateGeneralSettings(settings *general_settings.GeneralSettings) error {
	return gsr.db.DB().Table(gsr.generalSettingsModel.TableName()).
		Omit("id", "category", "name").
		Where("id = ?", settings.ID).
		Updates(settings).Error
}

// GetGeneralSettingsByID gets a general settings by ID
func (gsr *GeneralSettingsRepo) GetGeneralSettingsByID(id uint64) (*general_settings.GeneralSettings, error) {
	var settings general_settings.GeneralSettings
	err := gsr.db.DB().Table(gsr.generalSettingsModel.TableName()).
		Where("id = ?", id).
		First(&settings).Error
	return &settings, err
}

// GetGeneralSettingsByName gets a general settings by category and name
func (gsr *GeneralSettingsRepo) GetGeneralSettingsByName(category, name string) (*general_settings.GeneralSettings, error) {
	var settings general_settings.GeneralSettings
	err := gsr.db.DB().Table(gsr.generalSettingsModel.TableName()).
		Where("category = ? AND name = ?", category, name).
		First(&settings).Error
	return &settings, err
}

// GetGeneralSettingsList gets general settings list with pagination and filters
func (gsr *GeneralSettingsRepo) GetGeneralSettingsList(pageNum, pageSize int, category string) ([]*general_settings.GeneralSettings, int64, error) {
	var settingsList []*general_settings.GeneralSettings
	var total int64

	query := gsr.db.DB().Table(gsr.generalSettingsModel.TableName())

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
	err := gsr.db.DB().Table(gsr.generalSettingsModel.TableName()).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}
