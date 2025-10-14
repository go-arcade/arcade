package repo

import (
	"encoding/json"
	"fmt"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
)

type StorageRepo struct {
	Ctx          *ctx.Context
	StorageModel model.StorageConfig
}

func NewStorageRepo(ctx *ctx.Context) *StorageRepo {
	return &StorageRepo{
		Ctx:          ctx,
		StorageModel: model.StorageConfig{},
	}
}

// GetDefaultStorageConfig 获取默认存储配置
func (sr *StorageRepo) GetDefaultStorageConfig() (*model.StorageConfig, error) {
	var storageConfig model.StorageConfig
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("is_default = ? AND is_enabled = ?", 1, 1).
		First(&storageConfig).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get default storage config: %w", err)
	}
	return &storageConfig, nil
}

// GetStorageConfigByID 根据存储ID获取配置
func (sr *StorageRepo) GetStorageConfigByID(storageID string) (*model.StorageConfig, error) {
	var storageConfig model.StorageConfig
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("storage_id = ? AND is_enabled = ?", storageID, 1).
		First(&storageConfig).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get storage config by ID %s: %w", storageID, err)
	}
	return &storageConfig, nil
}

// GetEnabledStorageConfigs 获取所有启用的存储配置
func (sr *StorageRepo) GetEnabledStorageConfigs() ([]model.StorageConfig, error) {
	var storageConfigs []model.StorageConfig
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("is_enabled = ?", 1).
		Order("is_default DESC, id ASC").
		Find(&storageConfigs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled storage configs: %w", err)
	}
	return storageConfigs, nil
}

// GetStorageConfigByType 根据存储类型获取配置
func (sr *StorageRepo) GetStorageConfigByType(storageType string) ([]model.StorageConfig, error) {
	var storageConfigs []model.StorageConfig
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("storage_type = ? AND is_enabled = ?", storageType, 1).
		Order("is_default DESC, id ASC").
		Find(&storageConfigs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get storage configs by type %s: %w", storageType, err)
	}
	return storageConfigs, nil
}

// CreateStorageConfig 创建存储配置
func (sr *StorageRepo) CreateStorageConfig(storageConfig *model.StorageConfig) error {
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).Create(storageConfig).Error
	if err != nil {
		return fmt.Errorf("failed to create storage config: %w", err)
	}
	return nil
}

// UpdateStorageConfig 更新存储配置
func (sr *StorageRepo) UpdateStorageConfig(storageConfig *model.StorageConfig) error {
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("storage_id = ?", storageConfig.StorageId).
		Updates(storageConfig).Error
	if err != nil {
		return fmt.Errorf("failed to update storage config: %w", err)
	}
	return nil
}

// DeleteStorageConfig 删除存储配置
func (sr *StorageRepo) DeleteStorageConfig(storageID string) error {
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("storage_id = ?", storageID).
		Delete(&model.StorageConfig{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete storage config: %w", err)
	}
	return nil
}

// SetDefaultStorageConfig 设置默认存储配置
func (sr *StorageRepo) SetDefaultStorageConfig(storageID string) error {
	// 先取消所有默认配置
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("is_default = ?", 1).
		Update("is_default", 0).Error
	if err != nil {
		return fmt.Errorf("failed to clear default storage configs: %w", err)
	}

	// 设置新的默认配置
	err = sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("storage_id = ?", storageID).
		Update("is_default", 1).Error
	if err != nil {
		return fmt.Errorf("failed to set default storage config: %w", err)
	}
	return nil
}

// ParseStorageConfig 解析存储配置JSON为具体配置结构
func (sr *StorageRepo) ParseStorageConfig(storageConfig *model.StorageConfig) (interface{}, error) {
	switch storageConfig.StorageType {
	case "minio":
		var config model.MinIOConfig
		if err := json.Unmarshal(storageConfig.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to parse MinIO config: %w", err)
		}
		return &config, nil
	case "s3":
		var config model.S3Config
		if err := json.Unmarshal(storageConfig.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to parse S3 config: %w", err)
		}
		return &config, nil
	case "oss":
		var config model.OSSConfig
		if err := json.Unmarshal(storageConfig.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to parse OSS config: %w", err)
		}
		return &config, nil
	case "gcs":
		var config model.GCSConfig
		if err := json.Unmarshal(storageConfig.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to parse GCS config: %w", err)
		}
		return &config, nil
	case "cos":
		var config model.COSConfig
		if err := json.Unmarshal(storageConfig.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to parse COS config: %w", err)
		}
		return &config, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageConfig.StorageType)
	}
}
