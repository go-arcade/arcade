package minio

import (
	"mime/multipart"

	"github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/17 11:24
 * @file: minio.go
 * @description:
 */

type Minio struct {
	AccessKeyId     string
	SecretAccessKey string
	Endpoint        string
	Bucket          string
	UseTLS          bool
}

func NewMinio(accessKeyID, secretAccessKey, endpoint, bucket string, useTLS bool) *Minio {
	return &Minio{
		AccessKeyId:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Endpoint:        endpoint,
		Bucket:          bucket,
		UseTLS:          useTLS,
	}
}

func (m *Minio) Client() (*minio.Client, error) {
	minioClient, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKeyId, m.SecretAccessKey, ""),
		Secure: m.UseTLS,
	})
	if err != nil {
		log.Errorf("minio client error: %v", err)
	}
	return minioClient, nil
}

func (m *Minio) Upload(objectName string, file *multipart.FileHeader, contentType string, client minio.Client, ctx *fiber.Ctx) (minio.UploadInfo, error) {
	isExistBucket(ctx, client, m.Bucket)

	src, err := file.Open()
	if err != nil {
		return minio.UploadInfo{}, err
	}
	defer src.Close()

	result, err := client.PutObject(ctx.Context(), m.Bucket, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		_ = http.WithRepErrNotData(ctx, err.Error())
		return minio.UploadInfo{}, err
	}

	return result, nil
}

func (m *Minio) Download(objectName string, client minio.Client, ctx *fiber.Ctx) (*minio.Object, error) {
	isExistBucket(ctx, client, m.Bucket)
	obj, err := client.GetObject(ctx.Context(), m.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		_ = http.WithRepErrNotData(ctx, err.Error())
		return nil, err
	}
	return obj, nil
}

func isExistBucket(ctx *fiber.Ctx, client minio.Client, bucket string) {
	exists, err := client.BucketExists(ctx.Context(), bucket)
	if err != nil {
		return
	}
	if !exists {
		_ = http.WithRepErrNotData(ctx, "bucket not exists")
		return
	}
}
