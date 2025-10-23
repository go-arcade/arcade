package storage

import (
	"mime/multipart"
	"time"

	"github.com/go-arcade/arcade/pkg/ctx"
)

type StorageProvider interface {
	PutObject(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error)
	GetObject(ctx *ctx.Context, objectName string) ([]byte, error)
	Upload(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error)
	Download(ctx *ctx.Context, objectName string) ([]byte, error)
	Delete(ctx *ctx.Context, objectName string) error
	// GetPresignedURL 生成预签名下载链接，expiry 参数指定链接有效期
	GetPresignedURL(ctx *ctx.Context, objectName string, expiry time.Duration) (string, error)
}
