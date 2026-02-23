package material

import (
	"errors"
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/utils/material"
	"github.com/Project-DSView/backend/go/pkg/storage"
	"gorm.io/gorm"
)

// CodeExerciseService handles code exercise material operations
type CodeExerciseService struct {
	*BaseMaterialService
}

// NewCodeExerciseService creates a new code exercise service
func NewCodeExerciseService(db *gorm.DB, storageService storage.StorageInterface) *CodeExerciseService {
	return &CodeExerciseService{
		BaseMaterialService: NewBaseMaterialService(db, storageService),
	}
}

// CreateCodeExercise creates a new code exercise material
func (s *CodeExerciseService) CreateCodeExercise(exercise *models.CodeExercise, testCases []models.TestCase) error {
	// Validate course exists
	if err := s.ValidateCourseExists(exercise.CourseID); err != nil {
		return err
	}

	// Validate creator is teacher
	if err := s.ValidateCreatorIsTeacher(exercise.CreatedBy); err != nil {
		return err
	}

	// Validate code exercise fields
	if err := material.ValidateCodeExerciseFields(exercise.TotalPoints, exercise.ProblemStatement); err != nil {
		return err
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create code exercise
		if err := tx.Create(exercise).Error; err != nil {
			return fmt.Errorf("failed to create code exercise: %w", err)
		}

		// Create test cases
		for i := range testCases {
			testCases[i].MaterialID = &exercise.MaterialID
			testCases[i].MaterialType = "code_exercise"
			if err := tx.Create(&testCases[i]).Error; err != nil {
				return fmt.Errorf("failed to create test case %d: %w", i+1, err)
			}
		}

		return nil
	})
}

// GetCodeExerciseByID retrieves a code exercise by ID
func (s *CodeExerciseService) GetCodeExerciseByID(exerciseID string) (*models.CodeExercise, error) {
	var exercise models.CodeExercise
	if err := s.db.Preload("Creator").Preload("Course").Preload("TestCases").
		First(&exercise, "material_id = ?", exerciseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("code exercise not found")
		}
		return nil, err
	}
	return &exercise, nil
}

// GetCodeExercisesByCourse retrieves code exercises for a specific course
func (s *CodeExerciseService) GetCodeExercisesByCourse(courseID string, week *int, limit, offset int) ([]models.CodeExercise, int64, error) {
	var exercises []models.CodeExercise
	var total int64

	query := s.db.Model(&models.CodeExercise{}).Where("course_id = ?", courseID)

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

// UpdateCodeExercise updates an existing code exercise
func (s *CodeExerciseService) UpdateCodeExercise(exerciseID string, userID string, updates map[string]interface{}) error {
	// Validate ownership
	if err := s.ValidateMaterialOwnership(exerciseID, userID, "code_exercise"); err != nil {
		return err
	}

	// Update code exercise
	if err := s.db.Model(&models.CodeExercise{}).Where("material_id = ?", exerciseID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update code exercise: %w", err)
	}

	return nil
}

// DeleteCodeExercise deletes a code exercise
func (s *CodeExerciseService) DeleteCodeExercise(exerciseID string, userID string) error {
	// Validate ownership
	if err := s.ValidateMaterialOwnership(exerciseID, userID, "code_exercise"); err != nil {
		return err
	}

	// Delete code exercise (test cases will be deleted via CASCADE)
	if err := s.db.Delete(&models.CodeExercise{}, "material_id = ?", exerciseID).Error; err != nil {
		return fmt.Errorf("failed to delete code exercise: %w", err)
	}

	return nil
}

// GetTestCases retrieves test cases for a code exercise
func (s *CodeExerciseService) GetTestCases(exerciseID string) ([]models.TestCase, error) {
	var testCases []models.TestCase
	if err := s.db.Where("material_id = ? AND material_type = ?", exerciseID, "code_exercise").
		Find(&testCases).Error; err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}
	return testCases, nil
}

// AddTestCase adds a test case to a code exercise
func (s *CodeExerciseService) AddTestCase(exerciseID string, testCase *models.TestCase) error {
	// Check if exercise exists
	var exercise models.CodeExercise
	if err := s.db.First(&exercise, "material_id = ?", exerciseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("code exercise not found")
		}
		return err
	}

	// Set material ID and type
	testCase.MaterialID = &exerciseID
	testCase.MaterialType = "code_exercise"

	// Create test case
	if err := s.db.Create(testCase).Error; err != nil {
		return fmt.Errorf("failed to create test case: %w", err)
	}

	return nil
}

// UpdateTestCase updates an existing test case
func (s *CodeExerciseService) UpdateTestCase(testCaseID string, userID string, updates map[string]interface{}) error {
	// Check if test case exists and get exercise
	var testCase models.TestCase
	if err := s.db.First(&testCase, "test_case_id = ?", testCaseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("test case not found")
		}
		return err
	}

	if testCase.MaterialID == nil {
		return errors.New("test case is not associated with a material")
	}

	// Validate ownership
	if err := s.ValidateMaterialOwnership(*testCase.MaterialID, userID, "code_exercise"); err != nil {
		return err
	}

	// Update test case
	if err := s.db.Model(&testCase).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update test case: %w", err)
	}

	return nil
}

// DeleteTestCase deletes a test case
func (s *CodeExerciseService) DeleteTestCase(testCaseID string, userID string) error {
	// Check if test case exists and get exercise
	var testCase models.TestCase
	if err := s.db.First(&testCase, "test_case_id = ?", testCaseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("test case not found")
		}
		return err
	}

	if testCase.MaterialID == nil {
		return errors.New("test case is not associated with a material")
	}

	// Validate ownership
	if err := s.ValidateMaterialOwnership(*testCase.MaterialID, userID, "code_exercise"); err != nil {
		return err
	}

	// Delete test case
	if err := s.db.Delete(&testCase).Error; err != nil {
		return fmt.Errorf("failed to delete test case: %w", err)
	}

	return nil
}


















