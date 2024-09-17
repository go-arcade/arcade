package minio

import (
	"github.com/arcade/arcade/pkg/httpx"
	"github.com/arcade/arcade/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"mime/multipart"
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

func (m *Minio) Upload(objectName string, file *multipart.FileHeader, contentType string, client minio.Client, ctx *gin.Context) (minio.UploadInfo, error) {

	isExistBucket(ctx, client, m.Bucket)

	src, err := file.Open()
	if err != nil {
		return minio.UploadInfo{}, err
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			return
		}
	}(src)

	result, err := client.PutObject(ctx, m.Bucket, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		httpx.WithRepErrNotData(ctx, err.Error())
		return minio.UploadInfo{}, err
	}

	return result, nil
}

func (m *Minio) Download(objectName string, client minio.Client, ctx *gin.Context) (*minio.Object, error) {
	isExistBucket(ctx, client, m.Bucket)
	obj, err := client.GetObject(ctx, m.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		httpx.WithRepErrNotData(ctx, err.Error())
		return nil, err
	}
	return obj, nil
}

func isExistBucket(ctx *gin.Context, client minio.Client, bucket string) {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return
	}
	if !exists {
		httpx.WithRepErrNotData(ctx, "bucket not exists")
		return
	}
}
