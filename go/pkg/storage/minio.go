package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// DeleteFile deletes a file from MinIO storage
func (m *MinIOService) DeleteFile(ctx context.Context, url string) error {
	// Extract key from URL
	// Format: http://endpoint/bucket/key
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid MinIO URL format: %s", url)
	}

	// Get key (everything after the bucket part)
	key := strings.Join(parts[3:], "/")

	err := m.client.RemoveObject(ctx, m.config.BucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete MinIO object: %w", err)
	}

	return nil
}

// GetFileInfo retrieves information about a file from MinIO storage
func (m *MinIOService) GetFileInfo(ctx context.Context, url string) (interface{}, error) {
	// Extract key from URL
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid MinIO URL format: %s", url)
	}

	key := strings.Join(parts[3:], "/")

	info, err := m.client.StatObject(ctx, m.config.BucketName, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get MinIO object info: %w", err)
	}

	return &info, nil
}

// GeneratePresignedUploadURL generates a presigned URL for uploading files
func (m *MinIOService) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiration time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, m.config.BucketName, key, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GetFileURL generates a public URL for a file
func (m *MinIOService) GetFileURL(key string) string {
	return fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
}
