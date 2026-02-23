package services

import (
	"errors"
	"fmt"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"gorm.io/gorm"
)

type DeadlineCheckerService struct {
	db *gorm.DB
}

func NewDeadlineCheckerService(db *gorm.DB) *DeadlineCheckerService {
	return &DeadlineCheckerService{db: db}
}

// CheckMaterialDeadline checks if a course material is past its deadline
func (d *DeadlineCheckerService) CheckMaterialDeadline(materialID string) (bool, error) {
	var material models.CourseMaterial
	if err := d.db.First(&material, "material_id = ?", materialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("material not found")
		}
		return false, err
	}

	// Only exercises have deadlines
	if material.Type != enums.MaterialTypeCodeExercise && material.Type != enums.MaterialTypePDFExercise {
		return false, nil // Non-exercise materials don't have deadlines
	}

	// Get deadline from the actual exercise table
	var deadline *string
	if material.Type == enums.MaterialTypeCodeExercise {
		var codeExercise models.CodeExercise
		if err := d.db.First(&codeExercise, "material_id = ?", materialID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, errors.New("code exercise not found")
			}
			return false, err
		}
		deadline = codeExercise.Deadline
	} else if material.Type == enums.MaterialTypePDFExercise {
		var pdfExercise models.PDFExercise
		if err := d.db.First(&pdfExercise, "material_id = ?", materialID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, errors.New("PDF exercise not found")
			}
			return false, err
		}
		deadline = pdfExercise.Deadline
	}

	// If no deadline is set, material is always available
	if deadline == nil || *deadline == "" {
		return false, nil
	}

	// Parse deadline string to time
	deadlineTime, err := time.Parse(time.RFC3339, *deadline)
	if err != nil {
		return false, fmt.Errorf("invalid deadline format: %w", err)
	}

	// Check if current time is past deadline
	return time.Now().After(deadlineTime), nil
}

// GetAvailableMaterialsForUser gets materials that are available for a user (not past deadline)
func (d *DeadlineCheckerService) GetAvailableMaterialsForUser(userID string, courseID string) ([]models.CourseMaterial, error) {
	var materials []models.CourseMaterial

	// Get materials that are public and not past deadline
	query := d.db.Where("is_public = ?", true)

	// If courseID is provided, filter by course
	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}

	// Filter out materials past deadline
	query = query.Where("deadline IS NULL OR deadline > ?", time.Now().Format(time.RFC3339))

	if err := query.Preload("Creator").
		Order("created_at DESC").
		Find(&materials).Error; err != nil {
		return nil, err
	}

	return materials, nil
}

// GetExpiredMaterialsForUser gets materials that are past deadline for a user
func (d *DeadlineCheckerService) GetExpiredMaterialsForUser(userID string, courseID string) ([]models.CourseMaterial, error) {
	var materials []models.CourseMaterial

	// Get materials that are public and past deadline
	query := d.db.Where("is_public = ? AND deadline IS NOT NULL AND deadline <= ?",
		true, time.Now().Format(time.RFC3339))

	// If courseID is provided, filter by course
	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}

	if err := query.Preload("Creator").
		Order("deadline ASC").
		Find(&materials).Error; err != nil {
		return nil, err
	}

	return materials, nil
}

// CanSubmitExercise checks if a user can submit an exercise (deprecated)
func (d *DeadlineCheckerService) CanSubmitExercise(userID, exerciseID string) (bool, string, error) {
	return false, "Exercise submissions are deprecated. Please use course materials API instead.", nil
}

// CanSubmitMaterial checks if a user can submit a course material (PDF or code exercise)
func (d *DeadlineCheckerService) CanSubmitMaterial(userID, materialID string) (bool, string, error) {
	// Check if material exists (support both PDF and code exercises)
	var material models.CourseMaterial
	if err := d.db.First(&material, "material_id = ? AND (type = ? OR type = ?)",
		materialID, enums.MaterialTypePDFExercise, enums.MaterialTypeCodeExercise).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "Exercise not found", nil
		}
		return false, "", err
	}

	// Get the actual exercise to check deadline and is_public
	var deadline *string
	var isPublic bool

	if material.Type == enums.MaterialTypeCodeExercise {
		var codeExercise models.CodeExercise
		if err := d.db.First(&codeExercise, "material_id = ?", materialID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, "Code exercise not found", nil
			}
			return false, "", err
		}
		deadline = codeExercise.Deadline
		isPublic = codeExercise.IsPublic
	} else if material.Type == enums.MaterialTypePDFExercise {
		var pdfExercise models.PDFExercise
		if err := d.db.First(&pdfExercise, "material_id = ?", materialID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return false, "PDF exercise not found", nil
			}
			return false, "", err
		}
		deadline = pdfExercise.Deadline
		isPublic = pdfExercise.IsPublic
	}

	// Check if material is public
	if !isPublic {
		return false, "Material is not available", nil
	}

	// Check if material has deadline
	if deadline == nil || *deadline == "" {
		return true, "", nil
	}

	// Parse deadline string to time
	deadlineTime, err := time.Parse(time.RFC3339, *deadline)
	if err != nil {
		return false, "Invalid deadline format", nil
	}

	// Check if deadline has passed - block all submissions after deadline
	if time.Now().After(deadlineTime) {
		return false, "Submission deadline has passed", nil
	}

	return true, "", nil
}

// GetMaterialsByDeadlineStatus gets materials grouped by deadline status
func (d *DeadlineCheckerService) GetMaterialsByDeadlineStatus(courseID string) (map[string][]models.CourseMaterial, error) {
	result := make(map[string][]models.CourseMaterial)

	// Get all materials for the course
	var materials []models.CourseMaterial
	query := d.db.Model(&models.CourseMaterial{})

	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}

	if err := query.Find(&materials).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	for _, material := range materials {
		// Only exercises have deadlines
		if material.Type != enums.MaterialTypeCodeExercise && material.Type != enums.MaterialTypePDFExercise {
			// Non-exercise materials don't have deadlines - always available
			result["available"] = append(result["available"], material)
			continue
		}

		// Get deadline from the actual exercise table
		var deadline *string
		if material.Type == enums.MaterialTypeCodeExercise {
			var codeExercise models.CodeExercise
			if err := d.db.First(&codeExercise, "material_id = ?", material.MaterialID).Error; err != nil {
				continue // Skip if exercise not found
			}
			deadline = codeExercise.Deadline
		} else if material.Type == enums.MaterialTypePDFExercise {
			var pdfExercise models.PDFExercise
			if err := d.db.First(&pdfExercise, "material_id = ?", material.MaterialID).Error; err != nil {
				continue // Skip if exercise not found
			}
			deadline = pdfExercise.Deadline
		}

		if deadline == nil || *deadline == "" {
			// No deadline - always available
			result["available"] = append(result["available"], material)
		} else {
			// Parse deadline string to time
			deadlineTime, err := time.Parse(time.RFC3339, *deadline)
			if err != nil {
				continue // Skip invalid deadlines
			}

			if deadlineTime.After(now) {
				// Not yet expired
				result["available"] = append(result["available"], material)
			} else {
				// Expired
				result["expired"] = append(result["expired"], material)
			}
		}
	}

	return result, nil
}

// GetUpcomingDeadlines gets materials with deadlines approaching (within specified hours)
func (d *DeadlineCheckerService) GetUpcomingDeadlines(courseID string, hours int) ([]models.CourseMaterial, error) {
	var materials []models.CourseMaterial

	// Calculate cutoff time
	cutoffTime := time.Now().Add(time.Duration(hours) * time.Hour)

	query := d.db.Where("is_public = ? AND deadline IS NOT NULL AND deadline > ? AND deadline <= ?",
		true, time.Now().Format(time.RFC3339), cutoffTime.Format(time.RFC3339))

	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}

	if err := query.Preload("Creator").
		Order("deadline ASC").
		Find(&materials).Error; err != nil {
		return nil, err
	}

	return materials, nil
}

// GetDeadlineStats gets statistics about deadlines in a course
func (d *DeadlineCheckerService) GetDeadlineStats(courseID string) (map[string]interface{}, error) {
	var stats struct {
		TotalMaterials        int64
		MaterialsWithDeadline int64
		ExpiredMaterials      int64
		AvailableMaterials    int64
		UpcomingDeadlines     int64
	}

	// Count total public materials
	query := d.db.Model(&models.CourseMaterial{}).Where("is_public = ?", true)
	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}

	// Count total materials
	if err := query.Count(&stats.TotalMaterials).Error; err != nil {
		return nil, err
	}

	// Count materials with deadline
	if err := query.Where("deadline IS NOT NULL AND deadline != ''").Count(&stats.MaterialsWithDeadline).Error; err != nil {
		return nil, err
	}

	// Count expired materials
	now := time.Now().Format(time.RFC3339)
	if err := query.Where("deadline IS NOT NULL AND deadline != '' AND deadline <= ?", now).Count(&stats.ExpiredMaterials).Error; err != nil {
		return nil, err
	}

	// Count available materials
	if err := query.Where("deadline IS NULL OR deadline = '' OR deadline > ?", now).Count(&stats.AvailableMaterials).Error; err != nil {
		return nil, err
	}

	// Count upcoming deadlines (within 24 hours)
	cutoffTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	if err := query.Where("deadline IS NOT NULL AND deadline != '' AND deadline > ? AND deadline <= ?", now, cutoffTime).Count(&stats.UpcomingDeadlines).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_materials":         stats.TotalMaterials,
		"materials_with_deadline": stats.MaterialsWithDeadline,
		"expired_materials":       stats.ExpiredMaterials,
		"available_materials":     stats.AvailableMaterials,
		"upcoming_deadlines":      stats.UpcomingDeadlines,
	}, nil
}
