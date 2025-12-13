package storage

import (
	"context"
	"mime/multipart"
	"time"
)

type StorageProvider interface {
	PutObject(ctx context.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error)
	GetObject(ctx context.Context, objectName string) ([]byte, error)
	Upload(ctx context.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error)
	Download(ctx context.Context, objectName string) ([]byte, error)
	Delete(ctx context.Context, objectName string) error
	// GetPresignedURL 生成预签名下载链接，expiry 参数指定链接有效期
	GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}
