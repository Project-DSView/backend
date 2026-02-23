package services

import (
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"gorm.io/gorm"
)

type TestCaseService struct {
	db *gorm.DB
}

func NewTestCaseService(db *gorm.DB) *TestCaseService {
	return &TestCaseService{
		db: db,
	}
}

// TestCase operations

func (s *TestCaseService) CreateTestCase(testCase *models.TestCase) error {
	return s.db.Create(testCase).Error
}

func (s *TestCaseService) GetTestCasesByMaterialID(materialID string) ([]models.TestCase, error) {
	var testCases []models.TestCase
	if err := s.db.Where("material_id = ?", materialID).Order("created_at ASC").Find(&testCases).Error; err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}
	return testCases, nil
}

// GetTestCasesByMaterialIDWithFilter retrieves test cases for a material with optional public filter
func (s *TestCaseService) GetTestCasesByMaterialIDWithFilter(materialID string, publicOnly bool) ([]models.TestCase, error) {
	var testCases []models.TestCase

	query := s.db.Where("material_id = ?", materialID)

	if publicOnly {
		query = query.Where("is_public = ?", true)
	}

	if err := query.Order("created_at ASC").Find(&testCases).Error; err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}
	return testCases, nil
}

func (s *TestCaseService) UpdateTestCase(testCaseID string, updates map[string]interface{}) error {
	result := s.db.Model(&models.TestCase{}).Where("test_case_id = ?", testCaseID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update test case: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("test case not found")
	}
	return nil
}

func (s *TestCaseService) DeleteTestCase(testCaseID string) error {
	result := s.db.Where("test_case_id = ?", testCaseID).Delete(&models.TestCase{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete test case: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("test case not found")
	}
	return nil
}

func (s *TestCaseService) DeleteTestCasesByMaterialID(materialID string) error {
	return s.db.Where("material_id = ?", materialID).Delete(&models.TestCase{}).Error
}
