package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	Upload(ctx context.Context, reader io.Reader, size int64, contentType string) (key string, url string, err error)
	Get(ctx context.Context, key string) (reader io.ReadCloser, contentType string, err error)
	GetURL(key string) string
}

type S3Storage struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

// NewS3Storage creates an S3Storage. It no longer checks connectivity at
// construction time — the first real operation (Upload/Get) will surface
// connection errors. This lets the server start up cleanly even when MinIO
// credentials haven't been injected yet.
func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &S3Storage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

// ensureBucket lazily creates / verifies the bucket on first use.
func (s *S3Storage) ensureBucket(ctx context.Context) error {
	tctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	exists, err := s.client.BucketExists(tctx, s.bucket)
	if err != nil {
		return err
	}

	if !exists {
		if err := s.client.MakeBucket(tctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, s.bucket)

		if err := s.client.SetBucketPolicy(tctx, s.bucket, policy); err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Storage) Upload(ctx context.Context, reader io.Reader, size int64, contentType string) (string, string, error) {
	if err := s.ensureBucket(ctx); err != nil {
		return "", "", fmt.Errorf("storage unavailable: %w", err)
	}

	key := fmt.Sprintf("images/%s", uuid.New().String())

	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", err
	}

	url := s.GetURL(key)
	return key, url, nil
}

// Get streams an object from storage.
func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}

	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, "", err
	}

	return obj, info.ContentType, nil
}

// GetURL returns a relative URL served by the backend itself.
func (s *S3Storage) GetURL(key string) string {
	return "/api/" + key
}
