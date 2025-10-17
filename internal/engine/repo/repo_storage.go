package repo

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/observabil/arcade/internal/engine/model"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/log"
)

type StorageRepo struct {
	Ctx          *ctx.Context
	StorageModel model.StorageConfig
}

const (
	// Redis 缓存键前缀
	storageConfigCacheKeyPrefix  = "storage:config:"
	storageDefaultConfigCacheKey = "storage:config:default"

	// 缓存过期时间（1小时）
	storageConfigCacheTTL = 24 * time.Hour
)

func NewStorageRepo(ctx *ctx.Context) *StorageRepo {
	return &StorageRepo{
		Ctx:          ctx,
		StorageModel: model.StorageConfig{},
	}
}

// GetDefaultStorageConfig 获取默认存储配置（带Redis缓存）
func (sr *StorageRepo) GetDefaultStorageConfig() (*model.StorageConfig, error) {
	// 1. 先从 Redis 查询
	if sr.Ctx.RedisSession() != nil {
		cacheData, err := sr.Ctx.RedisSession().Get(sr.Ctx.ContextIns(), storageDefaultConfigCacheKey).Result()
		if err == nil && cacheData != "" {
			var storageConfig model.StorageConfig
			if err := json.Unmarshal([]byte(cacheData), &storageConfig); err == nil {
				log.Debugf("[StorageRepo] cache hit for default storage config")
				return &storageConfig, nil
			}
			log.Warnf("[StorageRepo] failed to unmarshal cached default storage config: %v", err)
		}
	}

	// 2. Redis 没有，从数据库查询
	var storageConfig model.StorageConfig
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("is_default = ? AND is_enabled = ?", 1, 1).
		First(&storageConfig).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get default storage config: %w", err)
	}

	// 3. 缓存到 Redis
	if sr.Ctx.RedisSession() != nil {
		cacheData, err := json.Marshal(storageConfig)
		if err == nil {
			err = sr.Ctx.RedisSession().Set(sr.Ctx.ContextIns(),
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
func (sr *StorageRepo) GetStorageConfigByID(storageID string) (*model.StorageConfig, error) {
	cacheKey := storageConfigCacheKeyPrefix + storageID

	// 1. 先从 Redis 查询
	if sr.Ctx.RedisSession() != nil {
		cacheData, err := sr.Ctx.RedisSession().Get(sr.Ctx.ContextIns(), cacheKey).Result()
		if err == nil && cacheData != "" {
			var storageConfig model.StorageConfig
			if err := json.Unmarshal([]byte(cacheData), &storageConfig); err == nil {
				log.Debugf("[StorageRepo] cache hit for storage config: %s", storageID)
				return &storageConfig, nil
			}
			log.Warnf("[StorageRepo] failed to unmarshal cached storage config %s: %v", storageID, err)
		}
	}

	// 2. Redis 没有，从数据库查询
	var storageConfig model.StorageConfig
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("storage_id = ? AND is_enabled = ?", storageID, 1).
		First(&storageConfig).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get storage config by ID %s: %w", storageID, err)
	}

	// 3. 缓存到 Redis
	if sr.Ctx.RedisSession() != nil {
		cacheData, err := json.Marshal(storageConfig)
		if err == nil {
			err = sr.Ctx.RedisSession().Set(sr.Ctx.ContextIns(),
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
	err := sr.Ctx.DBSession().Table(sr.StorageModel.TableName()).
		Where("storage_id = ?", storageID).
		Delete(&model.StorageConfig{}).Error
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

	// 清除默认配置缓存和相关配置缓存
	sr.clearDefaultStorageConfigCache()
	sr.clearStorageConfigCache(storageID)

	return nil
}

// clearStorageConfigCache 清除指定存储配置的缓存
func (sr *StorageRepo) clearStorageConfigCache(storageID string) {
	if sr.Ctx.RedisSession() == nil {
		return
	}

	cacheKey := storageConfigCacheKeyPrefix + storageID
	err := sr.Ctx.RedisSession().Del(sr.Ctx.ContextIns(), cacheKey).Err()
	if err != nil {
		log.Warnf("[StorageRepo] failed to clear cache for storage config %s: %v", storageID, err)
	} else {
		log.Debugf("[StorageRepo] cleared cache for storage config: %s", storageID)
	}
}

// clearDefaultStorageConfigCache 清除默认存储配置的缓存
func (sr *StorageRepo) clearDefaultStorageConfigCache() {
	if sr.Ctx.RedisSession() == nil {
		return
	}

	err := sr.Ctx.RedisSession().Del(sr.Ctx.ContextIns(), storageDefaultConfigCacheKey).Err()
	if err != nil {
		log.Warnf("[StorageRepo] failed to clear default storage config cache: %v", err)
	} else {
		log.Debugf("[StorageRepo] cleared default storage config cache")
	}
}

// ParseStorageConfig 解析存储配置JSON为具体配置结构
func (sr *StorageRepo) ParseStorageConfig(storageConfig *model.StorageConfig) (interface{}, error) {
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
		var config model.MinIOConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse MinIO config: %w", err)
		}
		return &config, nil
	case "s3":
		var config model.S3Config
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse S3 config: %w", err)
		}
		return &config, nil
	case "oss":
		var config model.OSSConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse OSS config: %w", err)
		}
		return &config, nil
	case "gcs":
		var config model.GCSConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse GCS config: %w", err)
		}
		return &config, nil
	case "cos":
		var config model.COSConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to parse COS config: %w", err)
		}
		return &config, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageConfig.StorageType)
	}
}
