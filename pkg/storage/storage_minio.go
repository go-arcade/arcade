package storage

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	Client *minio.Client
	s      *Storage
}

func newMinio(s *Storage) (*MinioStorage, error) {
	client, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.SecretKey, ""),
		Secure: s.UseTLS,
	})
	if err != nil {
		return nil, err
	}

	return &MinioStorage{
		Client: client,
		s:      s,
	}, nil
}

func (m *MinioStorage) GetObject(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(m.s.BasePath, objectName)
	obj, err := m.Client.GetObject(ctx.ContextIns(), m.s.Bucket, fullPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(obj)
	return buf.Bytes(), nil
}

func (m *MinioStorage) PutObject(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(m.s.BasePath, objectName)
	_, err = m.Client.PutObject(ctx.ContextIns(), m.s.Bucket, fullPath, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func (m *MinioStorage) Upload(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(m.s.BasePath, objectName)
	fileSize := file.Size

	// 小文件直接PutObject
	if fileSize <= defaultPartSize {
		_, err = m.Client.PutObject(ctx.ContextIns(), m.s.Bucket, fullPath, src, fileSize, minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err == nil {
			log.Debugf("MinIO upload completed: %s - 100.00%% (%d bytes)", fullPath, fileSize)
		}
		return fullPath, err
	}

	// 否则分片上传
	checkpointPath := filepath.Join(os.TempDir(), fullPath+".upload.json")
	var checkpoint uploadCheckpoint

	// 如果有断点记录则加载
	if data, err := os.ReadFile(checkpointPath); err == nil {
		_ = json.Unmarshal(data, &checkpoint)
	}

	// MinIO 会自动处理分片上传，我们只需要分块读取和记录进度
	if checkpoint.UploadID == "" {
		checkpoint = uploadCheckpoint{
			UploadID: fullPath, // MinIO 不需要显式的 UploadID
			Key:      fullPath,
			FileSize: fileSize,
		}
		_ = os.WriteFile(checkpointPath, mustJSON(checkpoint), 0644)
	}

	// 使用 PutObject，MinIO SDK 会自动处理大文件的分片上传
	// 使用 io.Reader 包装以支持断点续传
	partNumber := int32(1)
	buf := make([]byte, defaultPartSize)
	uploadedSize := int64(0)

	// 如果有已上传的分片，跳过它们
	for _, part := range checkpoint.Parts {
		if part >= partNumber {
			n, _ := src.Read(buf)
			if n > 0 {
				uploadedSize += int64(n)
				partNumber++
			}
		}
	}

	// 创建一个新的 reader 从当前位置开始
	remainingReader := newProgressReader(src, uploadedSize, fileSize, fullPath, "MinIO", func(uploaded int64) {
		currentPart := int32(uploaded / defaultPartSize)
		// 避免重复记录相同的分片
		if len(checkpoint.Parts) == 0 || checkpoint.Parts[len(checkpoint.Parts)-1] != currentPart {
			checkpoint.Parts = append(checkpoint.Parts, currentPart)
			checkpoint.UploadedBytes = uploaded
			checkpoint.UploadProgress = float64(uploaded) / float64(fileSize) * 100
			_ = os.WriteFile(checkpointPath, mustJSON(checkpoint), 0644)

			// 记录上传进度日志
			// log.Debugf("MinIO upload progress: %s - %.2f%% (%d/%d bytes)",
			// 	fullPath, checkpoint.UploadProgress, uploaded, fileSize)
		}
	})

	_, err = m.Client.PutObject(ctx.ContextIns(), m.s.Bucket, fullPath, remainingReader, fileSize-uploadedSize, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err == nil {
		log.Debugf("MinIO upload completed: %s - 100.00%% (%d bytes)", fullPath, fileSize)
		_ = os.Remove(checkpointPath) // 成功则删除断点文件
	}
	return fullPath, err
}

func (m *MinioStorage) Download(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(m.s.BasePath, objectName)
	obj, err := m.Client.GetObject(ctx.ContextIns(), m.s.Bucket, fullPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *MinioStorage) Delete(ctx *ctx.Context, objectName string) error {
	fullPath := getFullPath(m.s.BasePath, objectName)
	return m.Client.RemoveObject(ctx.ContextIns(), m.s.Bucket, fullPath, minio.RemoveObjectOptions{})
}

func (m *MinioStorage) GetPresignedURL(ctx *ctx.Context, objectName string, expiry time.Duration) (string, error) {
	fullPath := getFullPath(m.s.BasePath, objectName)

	reqParams := make(url.Values)
	presignedURL, err := m.Client.PresignedGetObject(ctx.ContextIns(), m.s.Bucket, fullPath, expiry, reqParams)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}
