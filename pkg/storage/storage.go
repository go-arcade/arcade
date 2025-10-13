package storage

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/wire"
	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/internal/engine/repo"
	"github.com/observabil/arcade/pkg/ctx"
)

// ProviderSet 提供存储相关的依赖
var ProviderSet = wire.NewSet(ProvideStorageFromDB)

// 存储类型常量
const (
	StorageMinio = "minio"
	StorageS3    = "s3"
	StorageOSS   = "oss"
	StorageGCS   = "gcs"
	StorageCOS   = "cos"
)

// Storage 存储配置结构
type Storage struct {
	Ctx       *ctx.Context
	Provider  string
	AccessKey string
	SecretKey string
	Endpoint  string
	Bucket    string
	Region    string
	UseTLS    bool
	BasePath  string
}

// StorageDBProvider 从数据库加载存储配置的提供者
type StorageDBProvider struct {
	ctx           *ctx.Context
	storageRepo   *repo.StorageRepo
	storageConfig *model.StorageConfig
}

// ProvideStorageFromDB 提供从数据库加载的存储实例
func ProvideStorageFromDB(ctx *ctx.Context, storageRepo *repo.StorageRepo) (StorageProvider, error) {
	dbProvider, err := NewStorageDBProvider(ctx, storageRepo)
	if err != nil {
		panic(err)
	}
	return dbProvider.GetStorageProvider()
}

// NewStorage 根据配置创建存储提供者实例
func NewStorage(s *Storage) (StorageProvider, error) {
	switch s.Provider {
	case StorageMinio:
		return newMinio(s)
	case StorageS3:
		return newS3(s)
	case StorageOSS:
		return newOSS(s)
	case StorageGCS:
		return newGCS(s)
	case StorageCOS:
		return newCOS(s)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", s.Provider)
	}
}

// NewStorageDBProvider 创建从数据库加载存储配置的提供者
func NewStorageDBProvider(ctx *ctx.Context, storageRepo *repo.StorageRepo) (*StorageDBProvider, error) {
	// 获取默认存储配置
	storageConfig, err := storageRepo.GetDefaultStorageConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get default storage config: %w", err)
	}

	return &StorageDBProvider{
		ctx:           ctx,
		storageRepo:   storageRepo,
		storageConfig: storageConfig,
	}, nil
}

// GetStorageProvider 获取存储提供者实例
func (sdp *StorageDBProvider) GetStorageProvider() (StorageProvider, error) {
	// 解析存储配置
	config, err := sdp.storageRepo.ParseStorageConfig(sdp.storageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse storage config: %w", err)
	}

	// 根据存储类型创建对应的存储实例
	switch sdp.storageConfig.StorageType {
	case "minio":
		minioConfig, ok := config.(*model.MinIOConfig)
		if !ok {
			return nil, fmt.Errorf("invalid MinIO config type")
		}
		return sdp.createMinIOStorage(minioConfig)
	case "s3":
		s3Config, ok := config.(*model.S3Config)
		if !ok {
			return nil, fmt.Errorf("invalid S3 config type")
		}
		return sdp.createS3Storage(s3Config)
	case "oss":
		ossConfig, ok := config.(*model.OSSConfig)
		if !ok {
			return nil, fmt.Errorf("invalid OSS config type")
		}
		return sdp.createOSSStorage(ossConfig)
	case "gcs":
		gcsConfig, ok := config.(*model.GCSConfig)
		if !ok {
			return nil, fmt.Errorf("invalid GCS config type")
		}
		return sdp.createGCSStorage(gcsConfig)
	case "cos":
		cosConfig, ok := config.(*model.COSConfig)
		if !ok {
			return nil, fmt.Errorf("invalid COS config type")
		}
		return sdp.createCOSStorage(cosConfig)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", sdp.storageConfig.StorageType)
	}
}

// createMinIOStorage 创建 MinIO 存储实例
func (sdp *StorageDBProvider) createMinIOStorage(config *model.MinIOConfig) (StorageProvider, error) {
	storage := &Storage{
		Provider:  StorageMinio,
		Endpoint:  config.Endpoint,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Bucket:    config.Bucket,
		Region:    config.Region,
		UseTLS:    config.UseTLS,
		BasePath:  config.BasePath,
	}
	return NewStorage(storage)
}

// createS3Storage 创建 S3 存储实例
func (sdp *StorageDBProvider) createS3Storage(config *model.S3Config) (StorageProvider, error) {
	storage := &Storage{
		Provider:  StorageS3,
		Endpoint:  config.Endpoint,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Bucket:    config.Bucket,
		Region:    config.Region,
		UseTLS:    config.UseTLS,
		BasePath:  config.BasePath,
	}
	return NewStorage(storage)
}

// createOSSStorage 创建 OSS 存储实例
func (sdp *StorageDBProvider) createOSSStorage(config *model.OSSConfig) (StorageProvider, error) {
	storage := &Storage{
		Provider:  StorageOSS,
		Endpoint:  config.Endpoint,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Bucket:    config.Bucket,
		Region:    config.Region,
		UseTLS:    config.UseTLS,
		BasePath:  config.BasePath,
	}
	return NewStorage(storage)
}

// createGCSStorage 创建 GCS 存储实例
func (sdp *StorageDBProvider) createGCSStorage(config *model.GCSConfig) (StorageProvider, error) {
	storage := &Storage{
		Provider:  StorageGCS,
		Endpoint:  config.Endpoint,
		AccessKey: config.AccessKey,
		Bucket:    config.Bucket,
		Region:    config.Region,
		BasePath:  config.BasePath,
	}
	return NewStorage(storage)
}

// createCOSStorage 创建 COS 存储实例
func (sdp *StorageDBProvider) createCOSStorage(config *model.COSConfig) (StorageProvider, error) {
	storage := &Storage{
		Provider:  StorageCOS,
		Endpoint:  config.Endpoint,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Bucket:    config.Bucket,
		Region:    config.Region,
		UseTLS:    config.UseTLS,
		BasePath:  config.BasePath,
	}
	return NewStorage(storage)
}

// GetStorageConfig 获取当前存储配置
func (sdp *StorageDBProvider) GetStorageConfig() *model.StorageConfig {
	return sdp.storageConfig
}

// RefreshStorageConfig 刷新存储配置（从数据库重新加载）
func (sdp *StorageDBProvider) RefreshStorageConfig() error {
	storageConfig, err := sdp.storageRepo.GetDefaultStorageConfig()
	if err != nil {
		return fmt.Errorf("failed to refresh storage config: %w", err)
	}
	sdp.storageConfig = storageConfig
	return nil
}

// GetStorageConfigByID 根据ID获取存储配置
func (sdp *StorageDBProvider) GetStorageConfigByID(storageID string) (*model.StorageConfig, error) {
	return sdp.storageRepo.GetStorageConfigByID(storageID)
}

// GetAllStorageConfigs 获取所有存储配置
func (sdp *StorageDBProvider) GetAllStorageConfigs() ([]model.StorageConfig, error) {
	return sdp.storageRepo.GetEnabledStorageConfigs()
}

// SwitchStorageConfig 切换存储配置
func (sdp *StorageDBProvider) SwitchStorageConfig(storageID string) error {
	storageConfig, err := sdp.storageRepo.GetStorageConfigByID(storageID)
	if err != nil {
		return fmt.Errorf("failed to get storage config by ID %s: %w", storageID, err)
	}

	// 设置为默认配置
	err = sdp.storageRepo.SetDefaultStorageConfig(storageID)
	if err != nil {
		return fmt.Errorf("failed to set default storage config: %w", err)
	}

	// 更新当前配置
	sdp.storageConfig = storageConfig
	return nil
}

// getFullPath 组合 BasePath 和 objectName，返回完整的对象路径
func getFullPath(basePath, objectName string) string {
	if basePath == "" {
		return objectName
	}
	// 清理路径，避免双斜杠
	basePath = strings.Trim(basePath, "/")
	objectName = strings.TrimPrefix(objectName, "/")
	return filepath.Join(basePath, objectName)
}
