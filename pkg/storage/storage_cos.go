package storage

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/observabil/arcade/pkg/ctx"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSStorage struct {
	Client *cos.Client
	s      *Storage
}

func newCOS(s *Storage) (*COSStorage, error) {
	// 解析 Endpoint，腾讯云 COS 需要 bucket URL
	u, err := url.Parse(s.Endpoint)
	if err != nil {
		return nil, err
	}

	// 如果 Endpoint 不包含 bucket，则添加
	if s.Bucket != "" && u.Host != "" {
		u, _ = url.Parse("https://" + s.Bucket + "." + u.Host)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.AccessKey,
			SecretKey: s.SecretKey,
		},
	})

	return &COSStorage{
		Client: client,
		s:      s,
	}, nil
}

func (c *COSStorage) GetObject(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(c.s.BasePath, objectName)
	resp, err := c.Client.Object.Get(context.Background(), fullPath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *COSStorage) PutObject(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := getFullPath(c.s.BasePath, objectName)
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	_, err = c.Client.Object.Put(context.Background(), fullPath, src, opt)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func (c *COSStorage) Upload(ctx *ctx.Context, objectName string, file *multipart.FileHeader, contentType string) (string, error) {
	return c.PutObject(ctx, objectName, file, contentType)
}

func (c *COSStorage) Download(ctx *ctx.Context, objectName string) ([]byte, error) {
	fullPath := getFullPath(c.s.BasePath, objectName)
	resp, err := c.Client.Object.Get(context.Background(), fullPath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *COSStorage) Delete(ctx *ctx.Context, objectName string) error {
	fullPath := getFullPath(c.s.BasePath, objectName)
	_, err := c.Client.Object.Delete(context.Background(), fullPath)
	return err
}
