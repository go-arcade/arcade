package storage

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"

	"cloud.google.com/go/storage"
	"github.com/observabil/arcade/pkg/ctx"
	"google.golang.org/api/option"
)

type GCSStorage struct {
	Client *storage.Client
	Bucket *storage.BucketHandle
	s      *Storage
}

func newGCS(s *Storage) (*GCSStorage, error) {
	var opts []option.ClientOption

	// 如果提供了 AccessKey，将其作为 credentials JSON 文件路径
	if s.AccessKey != "" {
		opts = append(opts, option.WithCredentialsFile(s.AccessKey))
	}

	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(s.Bucket)

	return &GCSStorage{
		Client: client,
		Bucket: bucket,
		s:      s,
	}, nil
}

func (g *GCSStorage) GetObject(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(g.s.BasePath, objectName)
	reader, err := g.Bucket.Object(fullPath).NewReader(ctx.ContextIns())
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *GCSStorage) PutObject(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(g.s.BasePath, objectName)
	writer := g.Bucket.Object(fullPath).NewWriter(ctx.ContextIns())
	writer.ContentType = contentType

	if _, err := io.Copy(writer, src); err != nil {
		writer.Close()
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	return fullPath, nil
}

func (g *GCSStorage) Upload(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	return g.PutObject(ctx, objectName, file, contentType)
}

func (g *GCSStorage) Download(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(g.s.BasePath, objectName)
	reader, err := g.Bucket.Object(fullPath).NewReader(ctx.ContextIns())
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (g *GCSStorage) Delete(ctx *ctx.Context, objectName string) error {
	fullPath := getFullPath(g.s.BasePath, objectName)
	return g.Bucket.Object(fullPath).Delete(ctx.ContextIns())
}
