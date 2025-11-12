package storage

import (
	"encoding/json"
	"fmt"

	"github.com/go-arcade/arcade/internal/engine/model/storage"
	storagerepo "github.com/go-arcade/arcade/internal/engine/repo/storage"
	"github.com/go-arcade/arcade/pkg/ctx"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2025/01/15
 * @file: service_storage.go
 * @description: storage configuration service
 */

type StorageService struct {
	ctx         *ctx.Context
	storageRepo storagerepo.IStorageRepository
}

func NewStorageService(ctx *ctx.Context, storageRepo storagerepo.IStorageRepository) *StorageService {
	return &StorageService{
		ctx:         ctx,
		storageRepo: storageRepo,
	}
}

// CreateStorageConfig 创建存储配置
func (ss *StorageService) CreateStorageConfig(req *CreateStorageConfigRequest) (*storage.StorageConfig, error) {
	// 验证配置
	if err := ss.validateStorageConfig(req.StorageType, req.Config); err != nil {
		return nil, fmt.Errorf("invalid storage config: %w", err)
	}

	storageConfig := &storage.StorageConfig{
		StorageId:   req.StorageId,
		Name:        req.Name,
		StorageType: req.StorageType,
		Config:      req.Config,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		IsEnabled:   1,
	}

	if err := ss.storageRepo.CreateStorageConfig(storageConfig); err != nil {
		return nil, fmt.Errorf("failed to create storage config: %w", err)
	}

	// 如果设置为默认配置，需要先取消其他默认配置
	if req.IsDefault == 1 {
		if err := ss.storageRepo.SetDefaultStorageConfig(req.StorageId); err != nil {
			return nil, fmt.Errorf("failed to set as default storage config: %w", err)
		}
	}

	return storageConfig, nil
}

// UpdateStorageConfig 更新存储配置
func (ss *StorageService) UpdateStorageConfig(req *UpdateStorageConfigRequest) (*storage.StorageConfig, error) {
	// 验证配置
	if err := ss.validateStorageConfig(req.StorageType, req.Config); err != nil {
		return nil, fmt.Errorf("invalid storage config: %w", err)
	}

	storageConfig := &storage.StorageConfig{
		StorageId:   req.StorageId,
		Name:        req.Name,
		StorageType: req.StorageType,
		Config:      req.Config,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		IsEnabled:   req.IsEnabled,
	}

	if err := ss.storageRepo.UpdateStorageConfig(storageConfig); err != nil {
		return nil, fmt.Errorf("failed to update storage config: %w", err)
	}

	// 如果设置为默认配置，需要先取消其他默认配置
	if req.IsDefault == 1 {
		if err := ss.storageRepo.SetDefaultStorageConfig(req.StorageId); err != nil {
			return nil, fmt.Errorf("failed to set as default storage config: %w", err)
		}
	}

	return storageConfig, nil
}

// GetStorageConfig 获取存储配置
func (ss *StorageService) GetStorageConfig(storageID string) (*storage.StorageConfig, error) {
	return ss.storageRepo.GetStorageConfigByID(storageID)
}

// GetDefaultStorageConfig 获取默认存储配置
func (ss *StorageService) GetDefaultStorageConfig() (*storage.StorageConfig, error) {
	return ss.storageRepo.GetDefaultStorageConfig()
}

// ListStorageConfigs 获取存储配置列表
func (ss *StorageService) ListStorageConfigs() ([]storage.StorageConfig, error) {
	return ss.storageRepo.GetEnabledStorageConfigs()
}

// DeleteStorageConfig 删除存储配置
func (ss *StorageService) DeleteStorageConfig(storageID string) error {
	return ss.storageRepo.DeleteStorageConfig(storageID)
}

// SetDefaultStorageConfig 设置默认存储配置
func (ss *StorageService) SetDefaultStorageConfig(storageID string) error {
	return ss.storageRepo.SetDefaultStorageConfig(storageID)
}

// validateStorageConfig 验证存储配置
func (ss *StorageService) validateStorageConfig(storageType string, configJSON []byte) error {
	switch storageType {
	case "minio":
		var config storage.MinIOConfig
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid MinIO config: %w", err)
		}
		return ss.validateMinIOConfig(&config)
	case "s3":
		var config storage.S3Config
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid S3 config: %w", err)
		}
		return ss.validateS3Config(&config)
	case "oss":
		var config storage.OSSConfig
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid OSS config: %w", err)
		}
		return ss.validateOSSConfig(&config)
	case "gcs":
		var config storage.GCSConfig
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid GCS config: %w", err)
		}
		return ss.validateGCSConfig(&config)
	case "cos":
		var config storage.COSConfig
		if err := json.Unmarshal(configJSON, &config); err != nil {
			return fmt.Errorf("invalid COS config: %w", err)
		}
		return ss.validateCOSConfig(&config)
	default:
		return fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// validateMinIOConfig 验证 MinIO 配置
func (ss *StorageService) validateMinIOConfig(config *storage.MinIOConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("MinIO endpoint is required")
	}
	if config.AccessKey == "" {
		return fmt.Errorf("MinIO access key is required")
	}
	if config.SecretKey == "" {
		return fmt.Errorf("MinIO secret key is required")
	}
	if config.Bucket == "" {
		return fmt.Errorf("MinIO bucket is required")
	}
	return nil
}

// validateS3Config 验证 S3 配置
func (ss *StorageService) validateS3Config(config *storage.S3Config) error {
	if config.Endpoint == "" {
		return fmt.Errorf("s3 endpoint is required")
	}
	if config.AccessKey == "" {
		return fmt.Errorf("s3 access key is required")
	}
	if config.SecretKey == "" {
		return fmt.Errorf("s3 secret key is required")
	}
	if config.Bucket == "" {
		return fmt.Errorf("s3 bucket is required")
	}
	return nil
}

// validateOSSConfig 验证 OSS 配置
func (ss *StorageService) validateOSSConfig(config *storage.OSSConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("OSS endpoint is required")
	}
	if config.AccessKey == "" {
		return fmt.Errorf("OSS access key is required")
	}
	if config.SecretKey == "" {
		return fmt.Errorf("OSS secret key is required")
	}
	if config.Bucket == "" {
		return fmt.Errorf("OSS bucket is required")
	}
	return nil
}

// validateGCSConfig 验证 GCS 配置
func (ss *StorageService) validateGCSConfig(config *storage.GCSConfig) error {
	if config.AccessKey == "" {
		return fmt.Errorf("GCS service account key is required")
	}
	if config.Bucket == "" {
		return fmt.Errorf("GCS bucket is required")
	}
	return nil
}

// validateCOSConfig 验证 COS 配置
func (ss *StorageService) validateCOSConfig(config *storage.COSConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("COS endpoint is required")
	}
	if config.AccessKey == "" {
		return fmt.Errorf("COS access key is required")
	}
	if config.SecretKey == "" {
		return fmt.Errorf("COS secret key is required")
	}
	if config.Bucket == "" {
		return fmt.Errorf("COS bucket is required")
	}
	return nil
}

type CreateStorageConfigRequest struct {
	StorageId   string `json:"storageId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	StorageType string `json:"storageType" binding:"required"`
	Config      []byte `json:"config" binding:"required"`
	Description string `json:"description"`
	IsDefault   int    `json:"isDefault"`
}

type UpdateStorageConfigRequest struct {
	StorageId   string `json:"storageId" binding:"required"`
	Name        string `json:"name"`
	StorageType string `json:"storageType"`
	Config      []byte `json:"config"`
	Description string `json:"description"`
	IsDefault   int    `json:"isDefault"`
	IsEnabled   int    `json:"isEnabled"`
}
