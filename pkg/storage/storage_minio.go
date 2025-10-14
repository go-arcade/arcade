package storage

import (
	"bytes"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/observabil/arcade/pkg/ctx"
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
	return m.PutObject(ctx, objectName, file, contentType)
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
