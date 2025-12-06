package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/log"
)

type S3Storage struct {
	Client *s3.Client
	s      *Storage
}

func newS3(s *Storage) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     s.AccessKey,
				SecretAccessKey: s.SecretKey,
			},
		}),
		config.WithBaseEndpoint(s.Endpoint),
		config.WithRegion(s.Region))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3Storage{Client: client, s: s}, nil
}

func (s *S3Storage) GetObject(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(s.s.BasePath, objectName)
	out, err := s.Client.GetObject(ctx.ContextIns(), &s3.GetObjectInput{
		Bucket: aws.String(s.s.Bucket),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(out.Body)
	return buf.Bytes(), nil
}

func (s *S3Storage) PutObject(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(s.s.BasePath, objectName)
	_, err = s.Client.PutObject(ctx.ContextIns(), &s3.PutObjectInput{
		Bucket:      aws.String(s.s.Bucket),
		Key:         aws.String(fullPath),
		Body:        src,
		ContentType: aws.String(contentType),
	})
	return fullPath, err
}

func (s *S3Storage) Upload(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(s.s.BasePath, objectName)
	fileSize := file.Size

	// 小文件直接PutObject
	if fileSize <= defaultPartSize {
		_, err = s.Client.PutObject(ctx.ContextIns(), &s3.PutObjectInput{
			Bucket:      aws.String(s.s.Bucket),
			Key:         aws.String(fullPath),
			Body:        src,
			ContentType: aws.String(contentType),
		})
		if err == nil {
			log.Debug("S3 upload completed: %s - 100.00%% (%d bytes)", fullPath, fileSize)
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

	if checkpoint.UploadID == "" {
		createResp, err := s.Client.CreateMultipartUpload(ctx.ContextIns(), &s3.CreateMultipartUploadInput{
			Bucket:      aws.String(s.s.Bucket),
			Key:         aws.String(fullPath),
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return "", err
		}
		checkpoint = uploadCheckpoint{
			UploadID: *createResp.UploadId,
			Key:      fullPath,
			FileSize: fileSize,
		}
		_ = os.WriteFile(checkpointPath, mustJSON(checkpoint), 0644)
	}

	var completedParts []s3types.CompletedPart
	partNumber := int32(1)
	buf := make([]byte, defaultPartSize)
	uploadedBytes := checkpoint.UploadedBytes // 从断点恢复已上传字节数

	for {
		n, readErr := src.Read(buf)
		if n == 0 {
			break
		}

		// 如果已上传过该分片则跳过
		if slices.Contains(checkpoint.Parts, partNumber) {
			uploadedBytes += int64(n)
			partNumber++
			continue
		}

		partOutput, err := s.Client.UploadPart(ctx.ContextIns(), &s3.UploadPartInput{
			Bucket:     aws.String(s.s.Bucket),
			Key:        aws.String(fullPath),
			PartNumber: &partNumber,
			UploadId:   aws.String(checkpoint.UploadID),
			Body:       bytes.NewReader(buf[:n]),
		})
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}

		completedParts = append(completedParts, s3types.CompletedPart{
			ETag:       partOutput.ETag,
			PartNumber: &partNumber,
		})
		checkpoint.Parts = append(checkpoint.Parts, partNumber)
		uploadedBytes += int64(n)

		// 更新进度信息
		checkpoint.UploadedBytes = uploadedBytes
		checkpoint.UploadProgress = float64(uploadedBytes) / float64(fileSize) * 100
		_ = os.WriteFile(checkpointPath, mustJSON(checkpoint), 0644)

		// 记录上传进度日志
		log.Debug("S3 upload progress: %s - %.2f%% (%d/%d bytes)",
			fullPath, checkpoint.UploadProgress, uploadedBytes, fileSize)

		partNumber++
		if readErr == io.EOF {
			break
		}
	}

	// 完成分片上传
	_, err = s.Client.CompleteMultipartUpload(ctx.ContextIns(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.s.Bucket),
		Key:      aws.String(fullPath),
		UploadId: aws.String(checkpoint.UploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err == nil {
		// log.Debug("S3 upload completed: %s - 100.00%% (%d bytes)", fullPath, fileSize)
		_ = os.Remove(checkpointPath) // 成功则删除断点文件
	}
	return fullPath, err
}

func (s *S3Storage) Download(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(s.s.BasePath, objectName)
	out, err := s.Client.GetObject(ctx.ContextIns(), &s3.GetObjectInput{
		Bucket: aws.String(s.s.Bucket),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(out.Body)
	return buf.Bytes(), nil
}

func (s *S3Storage) Delete(ctx *ctx.Context, objectName string) error {
	fullPath := getFullPath(s.s.BasePath, objectName)
	_, err := s.Client.DeleteObject(ctx.ContextIns(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.s.Bucket),
		Key:    aws.String(fullPath),
	})
	return err
}

func (s *S3Storage) GetPresignedURL(ctx *ctx.Context, objectName string, expiry time.Duration) (string, error) {
	fullPath := getFullPath(s.s.BasePath, objectName)

	presignClient := s3.NewPresignClient(s.Client)
	presignResult, err := presignClient.PresignGetObject(ctx.ContextIns(), &s3.GetObjectInput{
		Bucket: aws.String(s.s.Bucket),
		Key:    aws.String(fullPath),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}
