package storage

import (
	"mime/multipart"

	"github.com/observabil/arcade/pkg/ctx"
)

type StorageProvider interface {
	PutObject(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error)
	GetObject(ctx *ctx.Context, objectName string) ([]byte, error)
	Upload(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error)
	Download(ctx *ctx.Context, objectName string) ([]byte, error)
	Delete(ctx *ctx.Context, objectName string) error
}
