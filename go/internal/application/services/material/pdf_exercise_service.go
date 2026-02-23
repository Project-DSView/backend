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

// PDFExerciseService handles PDF exercise material operations
type PDFExerciseService struct {
	*BaseMaterialService
}

// NewPDFExerciseService creates a new PDF exercise service
func NewPDFExerciseService(db *gorm.DB, storageService storage.StorageInterface) *PDFExerciseService {
	return &PDFExerciseService{
		BaseMaterialService: NewBaseMaterialService(db, storageService),
	}
}

// CreatePDFExercise creates a new PDF exercise material
func (s *PDFExerciseService) CreatePDFExercise(exercise *models.PDFExercise) error {
	// Validate course exists
	if err := s.ValidateCourseExists(exercise.CourseID); err != nil {
		return err
	}

	// Validate creator is teacher
	if err := s.ValidateCreatorIsTeacher(exercise.CreatedBy); err != nil {
		return err
	}

	// Validate PDF exercise fields
	if err := material.ValidatePDFExerciseFields(exercise.TotalPoints, exercise.FileURL, exercise.FileName); err != nil {
		return err
	}

	// Create PDF exercise
	if err := s.db.Create(exercise).Error; err != nil {
		return fmt.Errorf("failed to create PDF exercise: %w", err)
	}

	return nil
}

// GetPDFExerciseByID retrieves a PDF exercise by ID
func (s *PDFExerciseService) GetPDFExerciseByID(exerciseID string) (*models.PDFExercise, error) {
	var exercise models.PDFExercise
	if err := s.db.Preload("Creator").Preload("Course").
		First(&exercise, "material_id = ?", exerciseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("PDF exercise not found")
		}
		return nil, err
	}
	return &exercise, nil
}

// GetPDFExercisesByCourse retrieves PDF exercises for a specific course
func (s *PDFExerciseService) GetPDFExercisesByCourse(courseID string, week *int, limit, offset int) ([]models.PDFExercise, int64, error) {
	var exercises []models.PDFExercise
	var total int64

	query := s.db.Model(&models.PDFExercise{}).Where("course_id = ?", courseID)

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
		Find(&exercises).Error; err != nil {
		return nil, 0, err
	}

	return exercises, total, nil
}

// UpdatePDFExercise updates an existing PDF exercise
func (s *PDFExerciseService) UpdatePDFExercise(exerciseID string, userID string, updates map[string]interface{}) error {
	// Validate ownership
	if err := s.ValidateMaterialOwnership(exerciseID, userID, "pdf_exercise"); err != nil {
		return err
	}

	// Update PDF exercise
	if err := s.db.Model(&models.PDFExercise{}).Where("material_id = ?", exerciseID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update PDF exercise: %w", err)
	}

	return nil
}

// DeletePDFExercise deletes a PDF exercise
func (s *PDFExerciseService) DeletePDFExercise(exerciseID string, userID string) error {
	// Get exercise to delete file
	var exercise models.PDFExercise
	if err := s.db.First(&exercise, "material_id = ?", exerciseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("PDF exercise not found")
		}
		return err
	}

	// Validate ownership
	if err := s.ValidateMaterialOwnership(exerciseID, userID, "pdf_exercise"); err != nil {
		return err
	}

	// Delete file from storage
	if exercise.FileURL != "" {
		ctx := context.Background()
		if err := s.DeleteFile(ctx, exercise.FileURL); err != nil {
			// Log error but don't fail the deletion
			logger.Warnf("Failed to delete file from storage: %v", err)
		}
	}

	// Delete PDF exercise
	if err := s.db.Delete(&exercise).Error; err != nil {
		return fmt.Errorf("failed to delete PDF exercise: %w", err)
	}

	return nil
}

// UploadPDFExerciseFile uploads a file for a PDF exercise
func (s *PDFExerciseService) UploadPDFExerciseFile(ctx context.Context, courseID, userID string, file io.Reader, filename, contentType string) (string, error) {
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

	key := fmt.Sprintf("course-materials/%s/pdf-exercises/%s", courseID, filename)
	url, err := storageService.UploadFile(ctx, key, file, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return url, nil
}
