package domain

import (
	"context"
	"io"
	"time"
)

type StorageAdapter interface {
	Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error
	PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
