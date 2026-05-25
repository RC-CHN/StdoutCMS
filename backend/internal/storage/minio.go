package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"stdoutcms/internal/config"
)

type S3 struct {
	client *minio.Client
	cfg    config.StorageConfig
}

func NewS3(cfg config.StorageConfig) (*S3, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}
	return &S3{client: client, cfg: cfg}, nil
}

// Upload puts a file into the bucket and returns the public CDN URL.
func (s *S3) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	opts := minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "public, max-age=31536000, immutable",
	}

	_, err := s.client.PutObject(ctx, s.cfg.Bucket, key, reader, size, opts)
	if err != nil {
		return "", fmt.Errorf("PutObject: %w", err)
	}

	// public URL = CDN_BASE_URL + key
	return s.cfg.CDNBaseURL + "/" + key, nil
}

// Delete removes a file from the bucket by key.
func (s *S3) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.cfg.Bucket, key, minio.RemoveObjectOptions{})
}

// GenerateKey returns an object key from a UUID and filename extension.
func GenerateKey(uuid, ext string) string {
	if ext == "" {
		ext = ".bin"
	}
	return "images/" + uuid + filepath.Ext(ext)
}
