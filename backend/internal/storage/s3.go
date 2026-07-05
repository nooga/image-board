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

func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create bucket if it doesn't exist
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}

		// Set bucket policy to allow public read access
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
		}`, bucket)

		err = client.SetBucketPolicy(ctx, bucket, policy)
		if err != nil {
			return nil, err
		}
	}

	return &S3Storage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
		useSSL:   useSSL,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, reader io.Reader, size int64, contentType string) (string, string, error) {
	// Generate unique key
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

// Get streams an object from storage. MinIO is private to the cluster, so
// clients never talk to it directly; the backend proxies image bytes instead.
func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}

	// GetObject is lazy; Stat forces the fetch so missing keys surface as errors.
	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, "", err
	}

	return obj, info.ContentType, nil
}

// GetURL returns a relative URL served by the backend itself (proxied publicly
// by nginx at /api/), rather than a direct MinIO endpoint that is only
// reachable inside the cluster.
func (s *S3Storage) GetURL(key string) string {
	return "/api/" + key
}

