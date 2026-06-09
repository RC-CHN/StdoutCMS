package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

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

// ListObjectKeys returns all object keys under the given prefix (recursive).
func (s *S3) ListObjectKeys(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	for obj := range s.client.ListObjects(ctx, s.cfg.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("ListObjects: %w", obj.Err)
		}
		keys = append(keys, obj.Key)
	}
	return keys, nil
}

// GenerateKey returns an object key from a UUID and a content-type (e.g. "image/png")
// or plain extension (e.g. ".jpg"). Falls back to ".bin" if neither is usable.
func GenerateKey(uuid, contentType string) string {
	ext := ""
	// Try MIME type first (e.g. "image/png" → ".png")
	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
		ext = exts[0]
	}
	// Fallback: treat contentType as a plain extension
	if ext == "" {
		if strings.HasPrefix(contentType, ".") {
			ext = contentType
		} else if contentType != "" {
			ext = "." + contentType
		}
	}
	if ext == "" {
		ext = ".bin"
	}
	return "images/" + uuid + filepath.Ext(ext)
}
