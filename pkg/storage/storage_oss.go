package storage

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/observabil/arcade/pkg/ctx"
	"github.com/observabil/arcade/pkg/log"
)

type OSSStorage struct {
	Client *oss.Client
	Bucket *oss.Bucket
	s      *Storage
}

func newOSS(s *Storage) (*OSSStorage, error) {
	client, err := oss.New(s.Endpoint, s.AccessKey, s.SecretKey)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(s.Bucket)
	if err != nil {
		return nil, err
	}

	return &OSSStorage{
		Client: client,
		Bucket: bucket,
		s:      s,
	}, nil
}

func (o *OSSStorage) GetObject(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(o.s.BasePath, objectName)
	body, err := o.Bucket.GetObject(fullPath)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *OSSStorage) PutObject(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(o.s.BasePath, objectName)
	err = o.Bucket.PutObject(fullPath, src, oss.ContentType(contentType))
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func (o *OSSStorage) Upload(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(o.s.BasePath, objectName)
	fileSize := file.Size

	// 小文件直接PutObject
	if fileSize <= defaultPartSize {
		err = o.Bucket.PutObject(fullPath, src, oss.ContentType(contentType))
		if err == nil {
			log.Infof("OSS upload completed: %s - 100.00%% (%d bytes)", fullPath, fileSize)
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

	// 初始化分片上传
	var imur oss.InitiateMultipartUploadResult
	if checkpoint.UploadID == "" {
		imur, err = o.Bucket.InitiateMultipartUpload(fullPath, oss.ContentType(contentType))
		if err != nil {
			return "", err
		}
		checkpoint = uploadCheckpoint{
			UploadID: imur.UploadID,
			Key:      fullPath,
			FileSize: fileSize,
		}
		_ = os.WriteFile(checkpointPath, mustJSON(checkpoint), 0644)
	}

	var parts []oss.UploadPart
	partNumber := 1
	buf := make([]byte, defaultPartSize)
	uploadedBytes := checkpoint.UploadedBytes // 从断点恢复已上传字节数

	for {
		n, readErr := src.Read(buf)
		if n == 0 {
			break
		}

		// 如果已上传过该分片则跳过
		skipPart := false
		for _, p := range checkpoint.Parts {
			if p == int32(partNumber) {
				skipPart = true
				break
			}
		}

		if skipPart {
			uploadedBytes += int64(n)
		} else {
			part, err := o.Bucket.UploadPart(imur, bytes.NewReader(buf[:n]), int64(n), partNumber)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
			checkpoint.Parts = append(checkpoint.Parts, int32(partNumber))
			uploadedBytes += int64(n)

			// 更新进度信息
			checkpoint.UploadedBytes = uploadedBytes
			checkpoint.UploadProgress = float64(uploadedBytes) / float64(fileSize) * 100
			_ = os.WriteFile(checkpointPath, mustJSON(checkpoint), 0644)

			// 记录上传进度日志
			// log.Debugf("OSS upload progress: %s - %.2f%% (%d/%d bytes)",
			// fullPath, checkpoint.UploadProgress, uploadedBytes, fileSize)
		}

		partNumber++
		if readErr == io.EOF {
			break
		}
	}

	// 完成分片上传
	_, err = o.Bucket.CompleteMultipartUpload(imur, parts)
	if err == nil {
		log.Debugf("OSS upload completed: %s - 100.00%% (%d bytes)", fullPath, fileSize)
		_ = os.Remove(checkpointPath) // 成功则删除断点文件
	}
	return fullPath, err
}

func (o *OSSStorage) Download(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(o.s.BasePath, objectName)
	body, err := o.Bucket.GetObject(fullPath)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	return io.ReadAll(body)
}

func (o *OSSStorage) Delete(ctx *ctx.Context, objectName string) error {
	fullPath := getFullPath(o.s.BasePath, objectName)
	return o.Bucket.DeleteObject(fullPath)
}

func (o *OSSStorage) GetPresignedURL(ctx *ctx.Context, objectName string, expiry time.Duration) (string, error) {
	fullPath := getFullPath(o.s.BasePath, objectName)

	// OSS 的过期时间以秒为单位
	expirySeconds := int64(expiry.Seconds())
	presignedURL, err := o.Bucket.SignURL(fullPath, oss.HTTPGet, expirySeconds)
	if err != nil {
		return "", err
	}

	return presignedURL, nil
}
