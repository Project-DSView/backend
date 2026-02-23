package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// UploadStudentPDFSubmission uploads a PDF file submitted by a student for an exercise
func (m *MinIOService) UploadStudentPDFSubmission(ctx context.Context, courseID, courseName string, week int, userEmail string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type - only allow PDF files
	if contentType != "application/pdf" {
		return "", fmt.Errorf("invalid file type: %s. Only PDF files are allowed", contentType)
	}

	// Build new path structure: {courseID-8}/exercise/{emailPrefix}/pdf/week-{N}/submission_timestamp.pdf
	coursePath := m.buildCoursePath(courseID)
	emailPrefix := m.buildStudentEmailPrefix(userEmail)
	weekPath := m.buildWeekPath(week)
	timestampFilename := m.buildTimestampFilename(filename)

	key := coursePath + "exercise/" + emailPrefix + "/pdf/" + weekPath + timestampFilename

	// Debug: log the key being used

	// Delete existing files in the same week directory to avoid duplicates
	weekDir := coursePath + "exercise/" + emailPrefix + "/pdf/" + weekPath
	m.deleteFilesInDirectory(ctx, weekDir)

	// Upload to MinIO
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"course-id":     courseID,
			"course-name":   courseName,
			"week":          fmt.Sprintf("%d", week),
			"user-email":    userEmail,
			"upload-type":   "student-pdf-submission",
			"original-name": filename,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload student PDF submission: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadStudentCodeSubmission uploads a code file submitted by a student for an exercise
func (m *MinIOService) UploadStudentCodeSubmission(ctx context.Context, courseID, courseName string, week int, userEmail string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type - allow common code file types
	if !m.isAllowedCodeType(contentType) {
		return "", fmt.Errorf("invalid code file type: %s. Allowed types: Python, JavaScript, Java, C++, C", contentType)
	}

	// Build new path structure: {courseID-8}/exercise/{emailPrefix}/code/week-{N}/submission_timestamp.{ext}
	coursePath := m.buildCoursePath(courseID)
	emailPrefix := m.buildStudentEmailPrefix(userEmail)
	weekPath := m.buildWeekPath(week)
	timestampFilename := m.buildTimestampFilename(filename)

	key := coursePath + "exercise/" + emailPrefix + "/code/" + weekPath + timestampFilename

	// Debug: log the key being used

	// Delete existing files in the same week directory to avoid duplicates
	weekDir := coursePath + "exercise/" + emailPrefix + "/code/" + weekPath
	m.deleteFilesInDirectory(ctx, weekDir)

	// Upload to MinIO
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"course-id":     courseID,
			"course-name":   courseName,
			"week":          fmt.Sprintf("%d", week),
			"user-email":    userEmail,
			"upload-type":   "student-code-submission",
			"original-name": filename,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload student code submission: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadStudentFeedbackFile uploads a feedback PDF file for a student submission
func (m *MinIOService) UploadStudentFeedbackFile(ctx context.Context, courseID, courseName string, week int, userEmail string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type - only allow PDF files
	if contentType != "application/pdf" {
		return "", fmt.Errorf("invalid file type: %s. Only PDF files are allowed", contentType)
	}

	// Build new path structure: {courseID-8}/exercise/{emailPrefix}/feedback/week-{N}/feedback_timestamp.pdf
	coursePath := m.buildCoursePath(courseID)
	emailPrefix := m.buildStudentEmailPrefix(userEmail)
	weekPath := m.buildWeekPath(week)
	feedbackFilename := m.buildFeedbackFilename(filename)

	key := coursePath + "exercise/" + emailPrefix + "/feedback/" + weekPath + feedbackFilename

	// Debug: log the key being used

	// Upload to MinIO (don't delete existing files, allow multiple feedback files)
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"course-id":     courseID,
			"course-name":   courseName,
			"week":          fmt.Sprintf("%d", week),
			"user-email":    userEmail,
			"upload-type":   "student-feedback",
			"original-name": filename,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload student feedback file: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// DownloadStudentPDFSubmission generates a presigned URL for downloading student PDF submissions
func (m *MinIOService) DownloadStudentPDFSubmission(ctx context.Context, fileURL string, expiration time.Duration) (string, error) {
	// Extract key from URL
	// URL format: http://{endpoint}/{bucketName}/{key}
	// For new path format: http://{endpoint}/{bucketName}/pdf/{materialID}/{userID}/{uniqueFilename}
	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid MinIO URL format: %s", fileURL)
	}

	// Find the bucket name in the URL and extract everything after it
	// parts[0] = "http:", parts[1] = "", parts[2] = "{endpoint}", parts[3] = "{bucketName}", parts[4+] = key parts
	bucketNameIndex := -1
	for i, part := range parts {
		if part == m.config.BucketName {
			bucketNameIndex = i
			break
		}
	}

	if bucketNameIndex == -1 || bucketNameIndex >= len(parts)-1 {
		// Fallback: assume bucket name is at index 3 (old format)
		// This handles both old and new URL formats
		if len(parts) > 4 {
			// Skip http:, empty, endpoint, bucketName - take everything after
			key := strings.Join(parts[4:], "/")
			presignedURL, err := m.client.PresignedGetObject(ctx, m.config.BucketName, key, expiration, nil)
			if err != nil {
				return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
			}
			return m.replacePresignedURLHost(presignedURL.String()), nil
		}
		return "", fmt.Errorf("invalid MinIO URL format: %s", fileURL)
	}

	// Extract key (everything after bucket name)
	key := strings.Join(parts[bucketNameIndex+1:], "/")

	// Generate presigned URL for download
	presignedURL, err := m.client.PresignedGetObject(ctx, m.config.BucketName, key, expiration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return m.replacePresignedURLHost(presignedURL.String()), nil
}

// StreamStudentPDFSubmission streams a PDF submission file from MinIO
func (m *MinIOService) StreamStudentPDFSubmission(ctx context.Context, fileURL string) (io.Reader, string, int64, error) {
	// Extract key from URL
	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 {
		return nil, "", 0, fmt.Errorf("invalid MinIO URL format: %s", fileURL)
	}

	// Find the bucket name in the URL and extract everything after it
	bucketNameIndex := -1
	for i, part := range parts {
		if part == m.config.BucketName {
			bucketNameIndex = i
			break
		}
	}

	if bucketNameIndex == -1 || bucketNameIndex >= len(parts)-1 {
		// Fallback: assume bucket name is at index 3 (old format)
		if len(parts) > 4 {
			key := strings.Join(parts[4:], "/")
			obj, err := m.client.GetObject(ctx, m.config.BucketName, key, minio.GetObjectOptions{})
			if err != nil {
				return nil, "", 0, fmt.Errorf("failed to get object: %w", err)
			}
			stat, err := obj.Stat()
			if err != nil {
				obj.Close()
				return nil, "", 0, fmt.Errorf("failed to stat object: %w", err)
			}
			return obj, stat.ContentType, stat.Size, nil
		}
		return nil, "", 0, fmt.Errorf("invalid MinIO URL format: %s", fileURL)
	}

	// Extract key (everything after bucket name)
	key := strings.Join(parts[bucketNameIndex+1:], "/")

	// Get object from MinIO
	obj, err := m.client.GetObject(ctx, m.config.BucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get object: %w", err)
	}

	// Get object info for content type and size
	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, "", 0, fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, stat.ContentType, stat.Size, nil
}

// replacePresignedURLHost replaces the host in presigned URL with public endpoint if configured
func (m *MinIOService) replacePresignedURLHost(presignedURLStr string) string {
	// If no public endpoint is configured, return original URL
	if m.config.PublicEndpoint == "" {
		return presignedURLStr
	}

	// Parse the presigned URL
	parsedURL, err := url.Parse(presignedURLStr)
	if err != nil {
		// If parsing fails, return original URL
		return presignedURLStr
	}

	// Replace host with public endpoint
	// PublicEndpoint format: "host:port" or "host" (e.g., "localhost:9000" or "minio.example.com")
	parsedURL.Host = m.config.PublicEndpoint

	// If public endpoint uses SSL but internal doesn't, update scheme
	// This is a simple check - if public endpoint contains common SSL indicators, use https
	// But we'll keep the original scheme for now unless explicitly configured
	// The UseSSL config should be checked, but for presigned URLs, we'll use the scheme from the original URL
	// unless PublicEndpoint starts with https://
	if strings.HasPrefix(m.config.PublicEndpoint, "https://") {
		parsedURL.Scheme = "https"
		parsedURL.Host = strings.TrimPrefix(m.config.PublicEndpoint, "https://")
	} else if strings.HasPrefix(m.config.PublicEndpoint, "http://") {
		parsedURL.Scheme = "http"
		parsedURL.Host = strings.TrimPrefix(m.config.PublicEndpoint, "http://")
	}

	return parsedURL.String()
}
