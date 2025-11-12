package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model/storage"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
)

type StorageRepo struct {
	db           database.DB
	cache        cache.Cache
	StorageModel storage.StorageConfig
}

const (
	// Redis 缓存键前缀
	storageConfigCacheKeyPrefix  = "storage:config:"
	storageDefaultConfigCacheKey = "storage:config:default"

	// 缓存过期时间（1小时）
	storageConfigCacheTTL = 24 * time.Hour
)

type IStorageRepository interface {
	GetDefaultStorageConfig() (*storage.StorageConfig, error)
	GetStorageConfigByID(storageID string) (*storage.StorageConfig, error)
	GetEnabledStorageConfigs() ([]storage.StorageConfig, error)
	GetStorageConfigByType(storageType string) ([]storage.StorageConfig, error)
	CreateStorageConfig(storageConfig *storage.StorageConfig) error
	UpdateStorageConfig(storageConfig *storage.StorageConfig) error
	DeleteStorageConfig(storageID string) error
	SetDefaultStorageConfig(storageID string) error
	ParseStorageConfig(storageConfig *storage.StorageConfig) (interface{}, error)
}

func NewStorageRepo(db database.DB, cache cache.Cache) IStorageRepository {
	return &StorageRepo{
		db:           db,
		cache:        cache,
		StorageModel: storage.StorageConfig{},
	}
}

// GetDB 返回数据库实例（供插件适配器使用）
func (sr *StorageRepo) GetDB() database.DB {
	return sr.db
}

// GetDefaultStorageConfig 获取默认存储配置（带Redis缓存）
func (sr *StorageRepo) GetDefaultStorageConfig() (*storage.StorageConfig, error) {
	// 1. 先从 Redis 查询
	ctx := context.Background()
	if sr.cache != nil {
		cacheData, err := sr.cache.Get(ctx, storageDefaultConfigCacheKey).Result()
		if err == nil && cacheData != "" {
			var storageConfig storage.StorageConfig
			if err := json.Unmarshal([]byte(cacheData), &storageConfig); err == nil {
				log.Debugf("[StorageRepo] cache hit for default storage config")
				return &storageConfig, nil
			}
			log.Warnf("[StorageRepo] failed to unmarshal cached default storage config: %v", err)
		}
	}

	// 2. Redis 没有，从数据库查询
	var storageConfig storage.StorageConfig
	err := sr.db.DB().Table(sr.StorageModel.TableName()).
		Where("is_default = ? AND is_enabled = ?", 1, 1).
		First(&storageConfig).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get default storage config: %w", err)
	}

	// 3. 缓存到 Redis
	if sr.cache != nil {
		cacheData, err := json.Marshal(storageConfig)
		if err == nil {
			err = sr.cache.Set(ctx,
				storageDefaultConfigCacheKey,
				cacheData,
				storageConfigCacheTTL).Err()
			if err != nil {
				log.Warnf("[StorageRepo] failed to cache default storage config: %v", err)
			} else {
				log.Debugf("[StorageRepo] cached default storage config")
			}
		}
	}

	return &storageConfig, nil
}

// GetStorageConfigByID 根据存储ID获取配置（带Redis缓存）
func (sr *StorageRepo) GetStorageConfigByID(storageID string) (*storage.StorageConfig, error) {
	cacheKey := storageConfigCacheKeyPrefix + storageID

	// 1. 先从 Redis 查询
	ctx := context.Background()
	if sr.cache != nil {
		cacheData, err := sr.cache.Get(ctx, cacheKey).Result()
		if err == nil && cacheData != "" {
			var storageConfig storage.StorageConfig
			if err := json.Unmarshal([]byte(cacheData), &storageConfig); err == nil {
				log.Debugf("[StorageRepo] cache hit for storage config: %s", storageID)
				return &storageConfig, nil
			}
			log.Warnf("[StorageRepo] failed to unmarshal cached storage config %s: %v", storageID, err)
		}
	}

	// 2. Redis 没有，从数据库查询
	var storageConfig storage.StorageConfig
	err := sr.db.DB().Table(sr.StorageModel.TableName()).
		Where("storage_id = ? AND is_enabled = ?", storageID, 1).
		First(&storageConfig).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get storage config by ID %s: %w", storageID, err)
	}

	// 3. 缓存到 Redis
	if sr.cache != nil {
		cacheData, err := json.Marshal(storageConfig)
		if err == nil {
			err = sr.cache.Set(ctx,
				cacheKey,
				cacheData,
				storageConfigCacheTTL).Err()
			if err != nil {
				log.Warnf("[StorageRepo] failed to cache storage config %s: %v", storageID, err)
			} else {
				log.Debugf("[StorageRepo] cached storage config: %s", storageID)
			}
		}
	}

	return &storageConfig, nil
}

// GetEnabledStorageConfigs 获取所有启用的存储配置
func (sr *StorageRepo) GetEnabledStorageConfigs() ([]storage.StorageConfig, error) {
	var storageConfigs []storage.StorageConfig
	err := sr.db.DB().Table(sr.StorageModel.TableName()).
		Select("storage_id", "name", "storage_type", "config", "description", "is_default", "is_enabled").
		Where("is_enabled = ?", 1).
		Order("is_default DESC, storage_id ASC").
		Find(&storageConfigs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled storage configs: %w", err)
	}
	return storageConfigs, nil
}

// GetStorageConfigByType 根据存储类型获取配置
func (sr *StorageRepo) GetStorageConfigByType(storageType string) ([]storage.StorageConfig, error) {
	var storageConfigs []storage.StorageConfig
	err := sr.db.DB().Table(sr.StorageModel.TableName()).
		Select("storage_id", "name", "storage_type", "config", "description", "is_default", "is_enabled").
		Where("storage_type = ? AND is_enabled = ?", storageType, 1).
		Order("is_default DESC, storage_id ASC").
		Find(&storageConfigs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get storage configs by type %s: %w", storageType, err)
	}
	return storageConfigs, nil
}

// CreateStorageConfig 创建存储配置
func (sr *StorageRepo) CreateStorageConfig(storageConfig *storage.StorageConfig) error {
	err := sr.db.DB().Table(sr.StorageModel.TableName()).Create(storageConfig).Error
	if err != nil {
		return fmt.Errorf("failed to create storage config: %w", err)
	}
	return nil
}

// UpdateStorageConfig 更新存储配置
func (sr *StorageRepo) UpdateStorageConfig(storageConfig *storage.StorageConfig) error {
	err := sr.db.DB().Table(sr.StorageModel.TableName()).
		Where("storage_id = ?", storageConfig.StorageId).
		Updates(storageConfig).Error
	if err != nil {
		return fmt.Errorf("failed to update storage config: %w", err)
	}

	// 清除缓存
	sr.clearStorageConfigCache(storageConfig.StorageId)
	// 如果是默认配置，也清除默认配置缓存
	if storageConfig.IsDefault == 1 {
		sr.clearDefaultStorageConfigCache()
	}

	return nil
}

// DeleteStorageConfig 删除存储配置
func (sr *StorageRepo) DeleteStorageConfig(storageID string) error {
	err := sr.db.DB().Table(sr.StorageModel.TableName()).
		Where("storage_id = ?", storageID).
		Delete(&storage.StorageConfig{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete storage config: %w", err)
	}

	// 清除缓存
	sr.clearStorageConfigCache(storageID)
	sr.clearDefaultStorageConfigCache()

	return nil
}

// SetDefaultStorageConfig 设置默认存储配置
func (sr *StorageRepo) SetDefaultStorageConfig(storageID string) error {
	// 先取消所有默认配置
	err := sr.db.DB().Table(sr.StorageModel.TableName()).
		Where("is_default = ?", 1).
		Update("is_default", 0).Error
	if err != nil {
		return fmt.Errorf("failed to clear default storage configs: %w", err)
	}

	// 设置新的默认配置
	err = sr.db.DB().Table(sr.StorageModel.TableName()).
		Where("storage_id = ?", storageID).
		Update("is_default", 1).Error
	if err != nil {
		return fmt.Errorf("failed to set default storage config: %w", err)
	}

	// 清除默认配置缓存和相关配置缓存
	sr.clearDefaultStorageConfigCache()
	sr.clearStorageConfigCache(storageID)

	return nil
}

// clearStorageConfigCache 清除指定存储配置的缓存
func (sr *StorageRepo) clearStorageConfigCache(storageID string) {
	if sr.cache == nil {
		return
	}
	ctx := context.Background()
	cacheKey := storageConfigCacheKeyPrefix + storageID
	err := sr.cache.Del(ctx, cacheKey).Err()
	if err != nil {
		log.Warnf("[StorageRepo] failed to clear cache for storage config %s: %v", storageID, err)
	} else {
		log.Debugf("[StorageRepo] cleared cache for storage config: %s", storageID)
	}
}

// clearDefaultStorageConfigCache 清除默认存储配置的缓存
func (sr *StorageRepo) clearDefaultStorageConfigCache() {
	if sr.cache == nil {
		return
	}
	ctx := context.Background()
	err := sr.cache.Del(ctx, storageDefaultConfigCacheKey).Err()
	if err != nil {
		log.Warnf("[StorageRepo] failed to clear default storage config cache: %v", err)
	} else {
		log.Debugf("[StorageRepo] cleared default storage config cache")
	}
}

// ParseStorageConfig 解析存储配置JSON为具体配置结构
func (sr *StorageRepo) ParseStorageConfig(storageConfig *storage.StorageConfig) (interface{}, error) {
	raw := storageConfig.Config

	// 兼容 Base64 + JSON 两种情况
	var configBytes []byte
	configStr := string(raw)
	if strings.HasPrefix(configStr, "{") {
		configBytes = raw
	} else {
		decoded, err := base64.StdEncoding.DecodeString(configStr)
		if err != nil {
			return nil, fmt.Errorf("invalid config encoding: %w", err)
		}
		configBytes = decoded
	}

	switch storageConfig.StorageType {
	case "minio":
		var config storage.MinIOConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse MinIO config: %w", err)
		}
		return &config, nil
	case "s3":
		var config storage.S3Config
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse S3 config: %w", err)
		}
		return &config, nil
	case "oss":
		var config storage.OSSConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse OSS config: %w", err)
		}
		return &config, nil
	case "gcs":
		var config storage.GCSConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse GCS config: %w", err)
		}
		return &config, nil
	case "cos":
		var config storage.COSConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse COS config: %w", err)
		}
		return &config, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageConfig.StorageType)
	}
}
