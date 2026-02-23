package storage

import (
	"context"
	"io"
	"time"
)

// StorageService interface สำหรับ object storage
type StorageService interface {
	// Code file operations
	UploadCodeFile(ctx context.Context, userID, exerciseID string, file io.Reader, filename, contentType string) (string, error)

	// PDF file operations
	UploadPDFFile(ctx context.Context, userID, materialID string, file io.Reader, filename, contentType string) (string, error)

	// Course image operations
	UploadCourseImage(ctx context.Context, courseID string, file io.Reader, filename, contentType string) (string, error)

	// Course material file operations
	UploadFile(ctx context.Context, key string, file io.Reader, contentType string) (string, error)
	UploadCourseMaterialFile(ctx context.Context, courseID, materialID, materialType string, file io.Reader, filename, contentType string) (string, error)
	UploadCourseVideo(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error)
	UploadCourseDocument(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error)
	UploadCourseCodeFile(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error)
	UploadCoursePDFFile(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error)

	// Problem image operations (for exercise problem statements)
	UploadProblemImage(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error)

	// Exercise-specific operations
	UploadExerciseCodeFile(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error)
	UploadExerciseImage(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error)

	// Student submission operations
	UploadStudentPDFSubmission(ctx context.Context, courseID, courseName string, week int, userEmail string, file io.Reader, filename, contentType string) (string, error)
	UploadStudentCodeSubmission(ctx context.Context, courseID, courseName string, week int, userEmail string, file io.Reader, filename, contentType string) (string, error)
	UploadStudentFeedbackFile(ctx context.Context, courseID, courseName string, week int, userEmail string, file io.Reader, filename, contentType string) (string, error)
	DownloadStudentPDFSubmission(ctx context.Context, fileURL string, expiration time.Duration) (string, error)
	StreamStudentPDFSubmission(ctx context.Context, fileURL string) (io.Reader, string, int64, error)

	// File management operations
	DeleteFile(ctx context.Context, url string) error
	GetFileInfo(ctx context.Context, url string) (interface{}, error)
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expiration time.Duration) (string, error)
	GetFileURL(key string) string

	// Utility operations
	HealthCheck(ctx context.Context) error
	ValidateFileSize(size int64) error
}

// StorageInterface is an alias for StorageService for backward compatibility
type StorageInterface = StorageService
