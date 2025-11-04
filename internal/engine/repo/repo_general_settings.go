package repo

import (
	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/15
 * @file: repo_general_settings.go
 * @description: general settings repository
 */

type GeneralSettingsRepo struct {
	ctx                  *ctx.Context
	generalSettingsModel *model.GeneralSettings
}

func NewGeneralSettingsRepo(ctx *ctx.Context) *GeneralSettingsRepo {
	return &GeneralSettingsRepo{
		ctx:                  ctx,
		generalSettingsModel: &model.GeneralSettings{},
	}
}

// UpdateGeneralSettings updates a general settings by ID
func (gsr *GeneralSettingsRepo) UpdateGeneralSettings(settings *model.GeneralSettings) error {
	return gsr.ctx.DBSession().Table(gsr.generalSettingsModel.TableName()).
		Omit("id", "category", "name").
		Where("id = ?", settings.ID).
		Updates(settings).Error
}

// GetGeneralSettingsByID gets a general settings by ID
func (gsr *GeneralSettingsRepo) GetGeneralSettingsByID(id uint64) (*model.GeneralSettings, error) {
	var settings model.GeneralSettings
	err := gsr.ctx.DBSession().Table(gsr.generalSettingsModel.TableName()).
		Where("id = ?", id).
		First(&settings).Error
	return &settings, err
}

// GetGeneralSettingsByName gets a general settings by category and name
func (gsr *GeneralSettingsRepo) GetGeneralSettingsByName(category, name string) (*model.GeneralSettings, error) {
	var settings model.GeneralSettings
	err := gsr.ctx.DBSession().Table(gsr.generalSettingsModel.TableName()).
		Where("category = ? AND name = ?", category, name).
		First(&settings).Error
	return &settings, err
}

// GetGeneralSettingsList gets general settings list with pagination and filters
func (gsr *GeneralSettingsRepo) GetGeneralSettingsList(pageNum, pageSize int, category string) ([]*model.GeneralSettings, int64, error) {
	var settingsList []*model.GeneralSettings
	var total int64

	query := gsr.ctx.DBSession().Table(gsr.generalSettingsModel.TableName())

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
	err := gsr.ctx.DBSession().Table(gsr.generalSettingsModel.TableName()).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}
