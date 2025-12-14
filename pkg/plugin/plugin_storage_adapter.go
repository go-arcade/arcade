// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"context"
)

// StorageProvider is the storage provider interface
// Consistent with the interface in the pkg/storage package
type StorageProvider interface {
	// Download downloads an object from the storage service
	Download(ctx *context.Context, objectName string) ([]byte, error)
}

// StorageAdapter is the storage adapter
// Implements the StorageDownloader interface, wrapping the storage service for plugin use
type StorageAdapter struct {
	// Storage service instance
	storage StorageProvider
	// Context
	ctx *context.Context
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(storage StorageProvider, ctx *context.Context) *StorageAdapter {
	return &StorageAdapter{
		storage: storage,
		ctx:     ctx,
	}
}

// Download implements the StorageDownloader interface
func (a *StorageAdapter) Download(objectName string) ([]byte, error) {
	return a.storage.Download(a.ctx, objectName)
}
