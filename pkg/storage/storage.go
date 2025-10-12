package storage

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/wire"
	"github.com/observabil/arcade/pkg/ctx"
)

// ProviderSet 提供存储相关的依赖
var ProviderSet = wire.NewSet(ProvideStorage)

// ProvideStorage 提供存储实例
func ProvideStorage(conf *Storage) (StorageProvider, error) {
	return NewStorage(conf)
}


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

const (
	StorageMinio = "minio"
	StorageS3    = "s3"
	StorageOSS   = "oss"
	StorageGCS   = "gcs"
	StorageCOS   = "cos"
)

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
