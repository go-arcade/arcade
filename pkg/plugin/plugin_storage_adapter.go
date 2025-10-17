// Package plugin storage adapter implementation
// Author: gagral.x@gmail.com
// Date: 2025/01/16
// Description: Storage service adapter that allows storage services to be used by the plugin manager
package plugin

import (
	"github.com/observabil/arcade/pkg/ctx"
)

// StorageProvider is the storage provider interface
// Consistent with the interface in the pkg/storage package
type StorageProvider interface {
	// Download downloads an object from the storage service
	Download(ctx *ctx.Context, objectName string) ([]byte, error)
}

// StorageAdapter is the storage adapter
// Implements the StorageDownloader interface, wrapping the storage service for plugin use
type StorageAdapter struct {
	// Storage service instance
	storage StorageProvider
	// Context
	ctx *ctx.Context
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(storage StorageProvider, ctx *ctx.Context) *StorageAdapter {
	return &StorageAdapter{
		storage: storage,
		ctx:     ctx,
	}
}

// Download implements the StorageDownloader interface
func (a *StorageAdapter) Download(objectName string) ([]byte, error) {
	return a.storage.Download(a.ctx, objectName)
}
