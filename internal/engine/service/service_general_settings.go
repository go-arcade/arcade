package service

import (
	"errors"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/log"
	"gorm.io/gorm"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/15
 * @file: service_general_settings.go
 * @description: general settings service
 */

type GeneralSettingsService struct {
	ctx                 *ctx.Context
	generalSettingsRepo repo.IGeneralSettingsRepository
}

func NewGeneralSettingsService(ctx *ctx.Context, generalSettingsRepo repo.IGeneralSettingsRepository) *GeneralSettingsService {
	return &GeneralSettingsService{
		ctx:                 ctx,
		generalSettingsRepo: generalSettingsRepo,
	}
}

// UpdateGeneralSettings updates a general settings
func (gss *GeneralSettingsService) UpdateGeneralSettings(id uint64, settings *model.GeneralSettings) error {
	// check if settings exists
	existing, err := gss.generalSettingsRepo.GetGeneralSettingsByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("general settings not found")
		}
		log.Errorf("failed to get general settings: %v", err)
		return errors.New("failed to get general settings")
	}

	// prevent changing category and name
	settings.ID = id
	settings.Category = existing.Category
	settings.Name = existing.Name

	if err := gss.generalSettingsRepo.UpdateGeneralSettings(settings); err != nil {
		log.Errorf("failed to update general settings: %v", err)
		return errors.New("failed to update general settings")
	}

	log.Infof("general settings updated successfully: %d", id)
	return nil
}

// GetGeneralSettingsByID gets a general settings by ID
func (gss *GeneralSettingsService) GetGeneralSettingsByID(id uint64) (*model.GeneralSettings, error) {
	settings, err := gss.generalSettingsRepo.GetGeneralSettingsByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("general settings not found")
		}
		log.Errorf("failed to get general settings: %v", err)
		return nil, errors.New("failed to get general settings")
	}
	return settings, nil
}

// GetGeneralSettingsByName gets a general settings by category and name
func (gss *GeneralSettingsService) GetGeneralSettingsByName(category, name string) (*model.GeneralSettings, error) {
	settings, err := gss.generalSettingsRepo.GetGeneralSettingsByName(category, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("general settings not found")
		}
		log.Errorf("failed to get general settings: %v", err)
		return nil, errors.New("failed to get general settings")
	}
	return settings, nil
}

// GetGeneralSettingsList gets general settings list with pagination and filters
func (gss *GeneralSettingsService) GetGeneralSettingsList(pageNum, pageSize int, category string) ([]*model.GeneralSettings, int64, error) {
	// set default pagination
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	settingsList, total, err := gss.generalSettingsRepo.GetGeneralSettingsList(pageNum, pageSize, category)
	if err != nil {
		log.Errorf("failed to get general settings list: %v", err)
		return nil, 0, errors.New("failed to get general settings list")
	}

	return settingsList, total, nil
}

// GetCategories gets all distinct categories
func (gss *GeneralSettingsService) GetCategories() ([]string, error) {
	categories, err := gss.generalSettingsRepo.GetCategories()
	if err != nil {
		log.Errorf("failed to get categories: %v", err)
		return nil, errors.New("failed to get categories")
	}
	return categories, nil
}
