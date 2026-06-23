package storage

import (
	"context"
	"io"
	"log/slog"
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
	slog.DebugContext(ctx, "s3: PutObject start",
		"bucket", a.bucket,
		"key", key,
		"size", size,
		"content_type", contentType,
	)
	info, err := a.client.PutObject(ctx, a.bucket, key, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		slog.ErrorContext(ctx, "s3: PutObject failed",
			"bucket", a.bucket,
			"key", key,
			"size", size,
			"err", err,
		)
		return err
	}
	slog.DebugContext(ctx, "s3: PutObject done",
		"bucket", a.bucket,
		"key", key,
		"etag", info.ETag,
		"version_id", info.VersionID,
	)
	return nil
}

func (a *S3Adapter) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	slog.DebugContext(ctx, "s3: PresignedGetObject start",
		"bucket", a.bucket,
		"key", key,
		"expiry", expiry,
	)
	u, err := a.client.PresignedGetObject(ctx, a.bucket, key, expiry, nil)
	if err != nil {
		slog.ErrorContext(ctx, "s3: PresignedGetObject failed",
			"bucket", a.bucket,
			"key", key,
			"err", err,
		)
		return "", err
	}
	slog.DebugContext(ctx, "s3: PresignedGetObject done",
		"bucket", a.bucket,
		"key", key,
		"url_host", u.Host,
	)
	return u.String(), nil
}

func (a *S3Adapter) Delete(ctx context.Context, key string) error {
	slog.DebugContext(ctx, "s3: RemoveObject",
		"bucket", a.bucket,
		"key", key,
	)
	err := a.client.RemoveObject(ctx, a.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "s3: RemoveObject failed",
			"bucket", a.bucket,
			"key", key,
			"err", err,
		)
	}
	return err
}
