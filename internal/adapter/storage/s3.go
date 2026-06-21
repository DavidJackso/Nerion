package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"nerion/internal/config"
	"nerion/internal/domain"
)

type S3Adapter struct {
	client *minio.Client
	bucket string
}

func NewS3Adapter(cfg config.StorageConfig) (domain.StorageAdapter, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: true,
		Region: cfg.S3Region,
	})
	if err != nil {
		return nil, err
	}
	return &S3Adapter{client: client, bucket: cfg.S3Bucket}, nil
}

func (a *S3Adapter) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	_, err := a.client.PutObject(ctx, a.bucket, key, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (a *S3Adapter) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := a.client.PresignedGetObject(ctx, a.bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (a *S3Adapter) Delete(ctx context.Context, key string) error {
	return a.client.RemoveObject(ctx, a.bucket, key, minio.RemoveObjectOptions{})
}
