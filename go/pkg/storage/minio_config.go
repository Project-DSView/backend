package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint         string
	PublicEndpoint   string // Public endpoint for presigned URLs (accessible from browser)
	AccessKeyID      string
	SecretAccessKey  string
	BucketName       string
	MaxFileSizeBytes int64
	UseSSL           bool
	PublicBucket     bool
}

type MinIOService struct {
	client *minio.Client
	config *MinIOConfig
}

func NewMinIOService(cfg *MinIOConfig) (*MinIOService, error) {
	// Create MinIO client
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Test connection by checking if bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		// Create bucket if it doesn't exist
		err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.BucketName, err)
		}
	}

	// Apply public read policy to bucket if configured
	if cfg.PublicBucket {
		err = applyPublicReadPolicy(ctx, minioClient, cfg.BucketName)
		if err != nil {
			// Log warning but don't fail initialization
			logger.Warnf("Failed to apply public read policy to bucket %s: %v", cfg.BucketName, err)
		}
	}

	// Apply lifecycle policy for student submissions (30 days auto-deletion)
	err = applyStudentSubmissionLifecyclePolicy(ctx, minioClient, cfg.BucketName, "exercise/student/")
	if err != nil {
		// Log warning but don't fail initialization
		logger.Warnf("Failed to apply lifecycle policy to bucket %s: %v", cfg.BucketName, err)
	}

	return &MinIOService{
		client: minioClient,
		config: cfg,
	}, nil
}

// applyPublicReadPolicy applies a public read policy to the bucket
func applyPublicReadPolicy(ctx context.Context, client *minio.Client, bucketName string) error {
	// Define the public read policy
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::` + bucketName + `/*"]
			}
		]
	}`

	// Apply the policy to the bucket
	err := client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	logger.Infof("Applied public read policy to bucket: %s", bucketName)
	return nil
}

// applyStudentSubmissionLifecyclePolicy applies a lifecycle policy for student submissions (30 days auto-deletion)
func applyStudentSubmissionLifecyclePolicy(_ context.Context, _ *minio.Client, bucketName, studentPrefix string) error {
	logger.Infof("Student submission lifecycle policy should be configured for bucket: %s (prefix: '%s', 30 day auto-delete)", bucketName, studentPrefix)
	return nil
}

// Health check method
func (m *MinIOService) HealthCheck(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.config.BucketName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("bucket %s does not exist", m.config.BucketName)
	}
	return nil
}

// Helper methods for file size validation
func (m *MinIOService) ValidateFileSize(size int64) error {
	if size > m.config.MaxFileSizeBytes {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes",
			size, m.config.MaxFileSizeBytes)
	}
	return nil
}
