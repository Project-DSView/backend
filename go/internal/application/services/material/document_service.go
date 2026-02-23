package material

import (
	"context"
	"errors"
	"fmt"
	"io"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/utils/material"
	"github.com/Project-DSView/backend/go/pkg/logger"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

// DocumentService handles document material operations
type DocumentService struct {
	*BaseMaterialService
}

// NewDocumentService creates a new document service
func NewDocumentService(db *gorm.DB, storageService storage.StorageInterface) *DocumentService {
	return &DocumentService{
		BaseMaterialService: NewBaseMaterialService(db, storageService),
	}
}

// CreateDocument creates a new document material
func (s *DocumentService) CreateDocument(document *models.Document) error {
	// Validate course exists
	if err := s.ValidateCourseExists(document.CourseID); err != nil {
		return err
	}

	// Validate creator is teacher
	if err := s.ValidateCreatorIsTeacher(document.CreatedBy); err != nil {
		return err
	}

	// Validate document fields
	if err := material.ValidateDocumentFields(document.FileURL, document.FileName); err != nil {
		return err
	}

	// Create document
	if err := s.db.Create(document).Error; err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// GetDocumentByID retrieves a document by ID
func (s *DocumentService) GetDocumentByID(documentID string) (*models.Document, error) {
	var document models.Document
	if err := s.db.Preload("Creator").Preload("Course").First(&document, "material_id = ?", documentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}
	return &document, nil
}

// GetDocumentsByCourse retrieves documents for a specific course
func (s *DocumentService) GetDocumentsByCourse(courseID string, week *int, limit, offset int) ([]models.Document, int64, error) {
	var documents []models.Document
	var total int64

	query := s.db.Model(&models.Document{}).Where("course_id = ?", courseID)

	if week != nil {
		query = query.Where("week = ?", *week)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Creator").
		Order("week ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// UpdateDocument updates an existing document
func (s *DocumentService) UpdateDocument(documentID string, userID string, updates map[string]interface{}) error {
	// Validate ownership
	if err := s.ValidateMaterialOwnership(documentID, userID, "document"); err != nil {
		return err
	}

	// Update document
	if err := s.db.Model(&models.Document{}).Where("material_id = ?", documentID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(documentID string, userID string) error {
	// Get document to delete file
	var document models.Document
	if err := s.db.First(&document, "material_id = ?", documentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("document not found")
		}
		return err
	}

	// Validate ownership
	if err := s.ValidateMaterialOwnership(documentID, userID, "document"); err != nil {
		return err
	}

	// Delete file from storage
	if document.FileURL != "" {
		ctx := context.Background()
		if err := s.DeleteFile(ctx, document.FileURL); err != nil {
			// Log error but don't fail the deletion
			logger.Warnf("Failed to delete file from storage: %v", err)
		}
	}

	// Delete document
	if err := s.db.Delete(&document).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// UploadDocumentFile uploads a file for a document
func (s *DocumentService) UploadDocumentFile(ctx context.Context, courseID, userID string, file io.Reader, filename, contentType string) (string, error) {
	// Validate user is teacher
	if err := s.ValidateCreatorIsTeacher(userID); err != nil {
		return "", err
	}

	// Validate course exists
	if err := s.ValidateCourseExists(courseID); err != nil {
		return "", err
	}

	// Upload to storage
	storageService := s.GetStorageService()
	if storageService == nil {
		return "", errors.New("storage service not available")
	}

	key := fmt.Sprintf("course-materials/%s/documents/%s", courseID, filename)
	url, err := storageService.UploadFile(ctx, key, file, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return url, nil
}
