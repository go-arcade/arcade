package storage

import (
	"bytes"
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/observabil/arcade/pkg/ctx"
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
	return s.PutObject(ctx, objectName, file, contentType)
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
