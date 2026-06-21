package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"nerion/internal/domain"
)

type LocalAdapter struct {
	uploadDir string
}

func NewLocalAdapter(uploadDir string) domain.StorageAdapter {
	return &LocalAdapter{uploadDir: uploadDir}
}

func (a *LocalAdapter) Upload(_ context.Context, key string, data io.Reader, _ int64, _ string) error {
	dest := filepath.Join(a.uploadDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, data)
	return err
}

// PresignedURL returns a local server path — not a real signed URL.
// Replace with S3 adapter for production use.
func (a *LocalAdapter) PresignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return fmt.Sprintf("/files/%s", key), nil
}

func (a *LocalAdapter) Delete(_ context.Context, key string) error {
	path := filepath.Join(a.uploadDir, filepath.FromSlash(key))
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
