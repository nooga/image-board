package storage

import (
	"context"
	"fmt"
	"io"
)

// NoopStorage is a no-op implementation of Storage used when MinIO is not
// configured. All operations return an error describing the missing config.
type NoopStorage struct{}

func NewNoopStorage() *NoopStorage {
	return &NoopStorage{}
}

func (n *NoopStorage) Upload(_ context.Context, _ io.Reader, _ int64, _ string) (string, string, error) {
	return "", "", fmt.Errorf("image storage not configured: set MINIO_ENDPOINT, MINIO_ACCESS_KEY, and MINIO_SECRET_KEY")
}

func (n *NoopStorage) Get(_ context.Context, _ string) (io.ReadCloser, string, error) {
	return nil, "", fmt.Errorf("image storage not configured")
}

func (n *NoopStorage) GetURL(key string) string {
	return "/api/" + key
}
