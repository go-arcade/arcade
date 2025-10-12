package storage

import (
	"bytes"
	"io"
	"mime/multipart"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/observabil/arcade/pkg/ctx"
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
	return o.PutObject(ctx, objectName, file, contentType)
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
