package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// UploadCodeFile uploads code files to object storage under codes/{exerciseID}/{userID}/
func (m *MinIOService) UploadCodeFile(ctx context.Context, userID, exerciseID string, file io.Reader, filename, contentType string) (string, error) {
	// Generate unique filename preserving extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".txt"
	}
	uniqueFilename := fmt.Sprintf("%s_%s_%d%s", userID, uuid.New().String()[:8], time.Now().Unix(), ext)
	key := "code/" + exerciseID + "/" + userID + "/" + uniqueFilename

	// Upload
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"uploaded-by": userID,
			"upload-type": "code-file",
			"exercise-id": exerciseID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload code file: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadPDFFile uploads PDF files to object storage under pdf/{materialID}/{userID}/
func (m *MinIOService) UploadPDFFile(ctx context.Context, userID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type - only allow PDF files
	if contentType != "application/pdf" && contentType != "application/x-pdf" {
		return "", fmt.Errorf("invalid file type: %s. Only PDF files are allowed", contentType)
	}

	// Generate unique filename preserving extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" || ext != ".pdf" {
		ext = ".pdf"
	}
	uniqueFilename := fmt.Sprintf("%s_%s_%d%s", userID, uuid.New().String()[:8], time.Now().Unix(), ext)
	key := "pdf/" + materialID + "/" + userID + "/" + uniqueFilename

	// Upload
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"uploaded-by":   userID,
			"upload-type":   "pdf-file",
			"material-id":   materialID,
			"original-name": filename,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF file: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadCourseImage uploads course images
func (m *MinIOService) UploadCourseImage(ctx context.Context, courseID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type
	if !m.isAllowedImageType(contentType) {
		return "", fmt.Errorf("invalid file type: %s. Allowed types: JPEG, PNG, WebP", contentType)
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	uniqueFilename := fmt.Sprintf("%s_%s_%d%s",
		courseID,
		uuid.New().String()[:8],
		time.Now().Unix(),
		ext)

	// Use course ID for path
	coursePath := m.buildCoursePath(courseID)
	key := coursePath + "image/" + uniqueFilename

	// Upload to MinIO
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"uploaded-for": courseID,
			"upload-type":  "course-image",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadFile uploads a file with a custom key
func (m *MinIOService) UploadFile(ctx context.Context, key string, file io.Reader, contentType string) (string, error) {
	// Upload to MinIO
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"upload-type": "course-material",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadCourseMaterialFile uploads course material files based on type
func (m *MinIOService) UploadCourseMaterialFile(ctx context.Context, courseID, materialID, materialType string, file io.Reader, filename, contentType string) (string, error) {
	var prefix string
	var uploadType string

	switch materialType {
	case "pdf_exercise", "document":
		prefix = "docs/"
		uploadType = "pdf-exercise"
	case "code_exercise":
		prefix = "code/"
		uploadType = "code-exercise"
	case "video":
		prefix = "videos/"
		uploadType = "video"
	default:
		prefix = "docs/"
		uploadType = "document"
	}

	// Generate unique filename preserving extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".txt"
	}

	// Create a simple, safe filename using only materialID, UUID, and timestamp
	uniqueFilename := fmt.Sprintf("file_%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)

	// Clean materialID to ensure it's safe for object keys
	cleanMaterialID := strings.ReplaceAll(materialID, "/", "_")
	cleanMaterialID = strings.ReplaceAll(cleanMaterialID, "\\", "_")

	// Handle empty materialID case
	if cleanMaterialID == "" {
		cleanMaterialID = "temp_" + uuid.New().String()[:8]
	}

	// Use course ID for path
	coursePath := m.buildCoursePath(courseID)
	key := coursePath + prefix + cleanMaterialID + "/" + uniqueFilename

	// Debug: log the key being used

	// Upload
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"course-id":     courseID,
			"material-id":   materialID,
			"material-type": materialType,
			"upload-type":   uploadType,
			"original-name": filename, // Store original filename in metadata
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload course material file: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadCourseVideo uploads video files for course materials
func (m *MinIOService) UploadCourseVideo(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type
	if !m.isAllowedVideoType(contentType) {
		return "", fmt.Errorf("invalid video file type: %s. Allowed types: MP4, WebM, AVI", contentType)
	}

	return m.UploadCourseMaterialFile(ctx, courseID, materialID, "video", file, filename, contentType)
}

// UploadCourseDocument uploads document files for course materials
func (m *MinIOService) UploadCourseDocument(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type
	if !m.isAllowedDocumentType(contentType) {
		return "", fmt.Errorf("invalid document file type: %s. Allowed types: PDF, DOC, DOCX, TXT", contentType)
	}

	return m.UploadCourseMaterialFile(ctx, courseID, materialID, "document", file, filename, contentType)
}

// UploadCourseCodeFile uploads code files for course materials
func (m *MinIOService) UploadCourseCodeFile(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type
	if !m.isAllowedCodeType(contentType) {
		return "", fmt.Errorf("invalid code file type: %s. Allowed types: Python, JavaScript, Java, C++, etc", contentType)
	}

	return m.UploadCourseMaterialFile(ctx, courseID, materialID, "code_exercise", file, filename, contentType)
}

// UploadCoursePDFFile uploads PDF and image files for course materials
func (m *MinIOService) UploadCoursePDFFile(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type - allow both PDF and image files
	if contentType != "application/pdf" && !m.isAllowedImageType(contentType) {
		return "", fmt.Errorf("invalid file type: %s. Only PDF files and images (JPEG, PNG, WebP) are allowed", contentType)
	}

	return m.UploadCourseMaterialFile(ctx, courseID, materialID, "pdf_exercise", file, filename, contentType)
}

// UploadExerciseCodeFile uploads code files for exercise templates/materials
func (m *MinIOService) UploadExerciseCodeFile(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type
	if !m.isAllowedCodeType(contentType) {
		return "", fmt.Errorf("invalid code file type: %s. Allowed types: Python, JavaScript, Java, C++, etc", contentType)
	}

	// Generate unique filename preserving extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".txt"
	}
	uniqueFilename := fmt.Sprintf("exercise_code_%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)

	// Clean materialID to ensure it's safe for object keys
	cleanMaterialID := strings.ReplaceAll(materialID, "/", "_")
	cleanMaterialID = strings.ReplaceAll(cleanMaterialID, "\\", "_")

	// Handle empty materialID case
	if cleanMaterialID == "" {
		cleanMaterialID = "temp_" + uuid.New().String()[:8]
	}

	// Use course ID for path
	coursePath := m.buildCoursePath(courseID)
	key := coursePath + "exercise/code/" + cleanMaterialID + "/" + uniqueFilename

	// Debug: log the key being used

	// Upload
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"course-id":     courseID,
			"material-id":   materialID,
			"upload-type":   "exercise-code",
			"original-name": filename,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload exercise code file: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadExerciseImage uploads images for exercise problem statements
func (m *MinIOService) UploadExerciseImage(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type - only allow image files
	if !m.isAllowedImageType(contentType) {
		return "", fmt.Errorf("invalid image file type: %s. Allowed types: JPEG, PNG, WebP, GIF", contentType)
	}

	// Generate unique filename preserving extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}
	uniqueFilename := fmt.Sprintf("exercise_img_%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)

	// Clean materialID to ensure it's safe for object keys
	cleanMaterialID := strings.ReplaceAll(materialID, "/", "_")
	cleanMaterialID = strings.ReplaceAll(cleanMaterialID, "\\", "_")

	// Handle empty materialID case
	if cleanMaterialID == "" {
		cleanMaterialID = "temp_" + uuid.New().String()[:8]
	}

	// Use course ID for path
	coursePath := m.buildCoursePath(courseID)
	key := coursePath + "exercise/image/" + cleanMaterialID + "/" + uniqueFilename

	// Debug: log the key being used

	// Upload
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"course-id":     courseID,
			"material-id":   materialID,
			"upload-type":   "exercise-image",
			"original-name": filename,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload exercise image: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}

// UploadProblemImage uploads images for exercise problem statements
func (m *MinIOService) UploadProblemImage(ctx context.Context, courseID, materialID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate file type - only allow image files
	if !m.isAllowedImageType(contentType) {
		return "", fmt.Errorf("invalid image file type: %s. Allowed types: JPEG, PNG, WebP, GIF", contentType)
	}

	// Generate unique filename preserving extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}
	uniqueFilename := fmt.Sprintf("problem_img_%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)

	// Clean materialID to ensure it's safe for object keys
	cleanMaterialID := strings.ReplaceAll(materialID, "/", "_")
	cleanMaterialID = strings.ReplaceAll(cleanMaterialID, "\\", "_")

	// Handle empty materialID case
	if cleanMaterialID == "" {
		cleanMaterialID = "temp_" + uuid.New().String()[:8]
	}

	// Use course ID for path
	coursePath := m.buildCoursePath(courseID)
	key := coursePath + "exercise/image/" + cleanMaterialID + "/" + uniqueFilename

	// Debug: log the key being used

	// Upload
	_, err := m.client.PutObject(ctx, m.config.BucketName, key, file, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"course-id":     courseID,
			"material-id":   materialID,
			"upload-type":   "problem-image",
			"original-name": filename,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload problem image: %w", err)
	}

	// Generate public URL
	url := fmt.Sprintf("http://%s/%s/%s", m.config.Endpoint, m.config.BucketName, key)
	return url, nil
}
