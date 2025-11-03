package service

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-arcade/arcade/internal/engine/model"
	"github.com/go-arcade/arcade/internal/engine/repo"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/id"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/storage"
)

type UploadService struct {
	ctx         *ctx.Context
	storageRepo *repo.StorageRepo
}

func NewUploadService(ctx *ctx.Context, storageRepo *repo.StorageRepo) *UploadService {
	return &UploadService{
		ctx:         ctx,
		storageRepo: storageRepo,
	}
}

const (
	userAvatarPath = "avatars"       // avatar storage path
	maxAvatarSize  = 5 * 1024 * 1024 // 5MB for avatar
)

var maxFileSize = int64(100 * 1024 * 1024) // 100MB for general files

// UploadFile uploads a file to object storage
// storageId: optional, if empty, use default storage
// path: optional custom path, if empty, use default path structure
func (us *UploadService) UploadFile(file *multipart.FileHeader, storageId string, customPath string) (*UploadResponse, error) {
	// validate file
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}

	if file.Size == 0 {
		return nil, fmt.Errorf("file size cannot be zero")
	}

	// get max file size (default 100MB)
	if file.Size > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of %d bytes", maxFileSize)
	}

	// get storage config
	var storageConfig *model.StorageConfig
	var err error
	if storageId != "" {
		storageConfig, err = us.storageRepo.GetStorageConfigByID(storageId)
		if err != nil {
			return nil, fmt.Errorf("failed to get storage config by ID: %w", err)
		}
	} else {
		storageConfig, err = us.storageRepo.GetDefaultStorageConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get default storage config: %w", err)
		}
	}

	// check if storage is enabled
	if storageConfig.IsEnabled != 1 {
		return nil, fmt.Errorf("storage config is disabled")
	}

	// create storage provider
	storageProvider, err := storage.NewStorageDBProvider(us.ctx, us.storageRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage provider: %w", err)
	}

	// if specific storage ID provided, switch to it
	if storageId != "" {
		if err := storageProvider.SwitchStorageConfig(storageId); err != nil {
			return nil, fmt.Errorf("failed to switch storage config: %w", err)
		}
	}

	// get storage provider instance
	provider, err := storageProvider.GetStorageProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage provider: %w", err)
	}

	// generate object name
	objectName := us.generateObjectName(file.Filename, customPath)

	// get content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = getContentType(file.Filename)
	}

	// upload file
	fullPath, err := provider.PutObject(us.ctx, objectName, file, contentType)
	if err != nil {
		log.Errorf("failed to upload file: %v", err)
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	log.Infof("file uploaded successfully: %s, size: %d", fullPath, file.Size)

	// generate presigned URL for download (optional, 7 days expiry)
	var downloadURL string
	if presignedProvider, ok := provider.(interface {
		GetPresignedURL(ctx *ctx.Context, objectName string, expiry time.Duration) (string, error)
	}); ok {
		downloadURL, _ = presignedProvider.GetPresignedURL(us.ctx, fullPath, 7*24*time.Hour)
	}

	return &UploadResponse{
		ObjectName:   fullPath,
		OriginalName: file.Filename,
		Size:         file.Size,
		ContentType:  contentType,
		StorageId:    storageConfig.StorageId,
		StorageType:  storageConfig.StorageType,
		DownloadURL:  downloadURL,
		UploadTime:   time.Now(),
	}, nil
}

// generateObjectName generates a unique object name
func (us *UploadService) generateObjectName(originalName string, customPath string) string {
	// get file extension
	ext := filepath.Ext(originalName)

	// generate unique ID for filename
	uniqueId := id.GetUUID()

	// sanitize original name (remove extension)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)
	nameWithoutExt = sanitizeFileName(nameWithoutExt)

	// construct filename
	fileName := fmt.Sprintf("%s_%s%s", nameWithoutExt, uniqueId, ext)

	// construct full path
	if customPath != "" {
		// use custom path
		customPath = strings.Trim(customPath, "/")
		return fmt.Sprintf("%s/%s", customPath, fileName)
	}

	// use default path: uploads/YYYY/MM/DD/filename
	now := time.Now()
	datePath := fmt.Sprintf("uploads/%04d/%02d/%02d", now.Year(), now.Month(), now.Day())
	return fmt.Sprintf("%s/%s", datePath, fileName)
}

// sanitizeFileName removes invalid characters from filename
func sanitizeFileName(name string) string {
	// replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// remove invalid characters
	invalidChars := []string{"/", "\\", "..", "<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "")
	}

	// limit length
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

// getContentType gets content type from file extension
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".zip":  "application/zip",
		".txt":  "text/plain",
		".json": "application/json",
		".xml":  "application/xml",
		".csv":  "text/csv",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}

// UploadAvatar uploads user avatar image
func (us *UploadService) UploadAvatar(file *multipart.FileHeader, userId string) (*UploadResponse, error) {
	// validate file
	if file == nil {
		return nil, fmt.Errorf("file is required")
	}

	// validate file size (max 5MB for avatar)
	if file.Size > maxAvatarSize {
		return nil, fmt.Errorf("avatar size exceeds maximum limit of 5MB")
	}

	// validate file type (only images)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = getContentType(file.Filename)
	}

	allowedTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp"}
	isValidType := false
	for _, t := range allowedTypes {
		if contentType == t {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return nil, fmt.Errorf("invalid image type, only jpeg, png, gif, and webp are allowed")
	}

	// generate custom path for avatar: avatars/userId
	customPath := fmt.Sprintf("%s/%s", userAvatarPath, userId)

	// upload file using default storage
	return us.UploadFile(file, "", customPath)
}

// UploadResponse response for file upload
type UploadResponse struct {
	ObjectName   string    `json:"objectName"`
	OriginalName string    `json:"originalName"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"contentType"`
	StorageId    string    `json:"storageId"`
	StorageType  string    `json:"storageType"`
	DownloadURL  string    `json:"downloadUrl,omitempty"`
	UploadTime   time.Time `json:"uploadTime"`
}
