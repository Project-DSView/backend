package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/minio/minio-go/v7"
)

// Helper methods for file type validation
func (m *MinIOService) isAllowedImageType(contentType string) bool {
	allowedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func (m *MinIOService) isAllowedVideoType(contentType string) bool {
	allowedTypes := []string{
		"video/mp4",
		"video/webm",
		"video/avi",
		"video/quicktime",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func (m *MinIOService) isAllowedDocumentType(contentType string) bool {
	allowedTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
		"text/html",
		"application/rtf",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func (m *MinIOService) isAllowedCodeType(contentType string) bool {
	allowedTypes := []string{
		"text/plain",
		"text/x-python",
		"text/javascript",
		"text/x-java-source",
		"text/x-c++src",
		"text/x-csrc",
		"text/x-go",
		"text/x-ruby",
		"text/x-php",
		"application/json",
		"text/html",
		"text/css",
		"text/xml",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// Helper methods for building paths

// buildCoursePath creates: {shortID}/
func (m *MinIOService) buildCoursePath(courseID string) string {
	cleanCourseID := strings.ReplaceAll(courseID, "/", "_")
	cleanCourseID = strings.ReplaceAll(cleanCourseID, "\\", "_")

	// Get first 8 characters of course ID
	shortCourseID := cleanCourseID
	if len(cleanCourseID) > 8 {
		shortCourseID = cleanCourseID[:8]
	}

	return shortCourseID + "/"
}

// buildStudentEmailPrefix extracts email prefix (before @)
func (m *MinIOService) buildStudentEmailPrefix(email string) string {
	// Extract part before @ and clean it
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return strings.ReplaceAll(parts[0], ".", "_")
	}
	return strings.ReplaceAll(email, ".", "_")
}

// buildWeekPath creates: week-{N}/
func (m *MinIOService) buildWeekPath(week int) string {
	return fmt.Sprintf("week-%d/", week)
}

// buildTimestampFilename creates: submission_YYYYMMDD_HHMMSS.{ext}
func (m *MinIOService) buildTimestampFilename(originalFilename string) string {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".pdf" // default for submissions
	}

	now := time.Now()
	return fmt.Sprintf("submission_%s_%s%s",
		now.Format("20060102"),
		now.Format("150405"),
		ext)
}

// buildFeedbackFilename creates: feedback_YYYYMMDD_HHMMSS.{ext}
func (m *MinIOService) buildFeedbackFilename(originalFilename string) string {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".pdf" // default for feedback files
	}

	now := time.Now()
	return fmt.Sprintf("feedback_%s_%s%s",
		now.Format("20060102"),
		now.Format("150405"),
		ext)
}

// deleteFilesInDirectory deletes all files in a specific directory
func (m *MinIOService) deleteFilesInDirectory(ctx context.Context, directoryPath string) {
	// List objects with the directory prefix
	objectCh := m.client.ListObjects(ctx, m.config.BucketName, minio.ListObjectsOptions{
		Prefix:    directoryPath,
		Recursive: true,
	})

	// Delete each object
	for object := range objectCh {
		if object.Err != nil {
			logger.Errorf("Error listing object: %v", object.Err)
			continue
		}

		err := m.client.RemoveObject(ctx, m.config.BucketName, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			logger.Errorf("Error deleting object %s: %v", object.Key, err)
		} else {

		}
	}
}
