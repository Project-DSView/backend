package material

import (
	"errors"
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"gorm.io/gorm"
)

// GetMaterialWithDetails retrieves the actual material data from the specific table and combines it with CourseMaterial reference
func GetMaterialWithDetails(db *gorm.DB, material *models.CourseMaterial) (map[string]interface{}, error) {
	// If no reference, return basic CourseMaterial data
	if material.ReferenceID == nil || material.ReferenceType == nil {
		result := material.ToJSON()
		return result, nil
	}

	referenceID := *material.ReferenceID
	referenceType := *material.ReferenceType

	// Query the specific material table based on reference_type
	switch referenceType {
	case "code_exercise":
		var codeExercise models.CodeExercise
		if err := db.Preload("Creator").Preload("TestCases").First(&codeExercise, "material_id = ?", referenceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// If not found, return basic CourseMaterial data
				return material.ToJSON(), nil
			}
			return nil, fmt.Errorf("failed to get code exercise: %w", err)
		}
		return codeExercise.ToJSON(), nil

	case "pdf_exercise":
		var pdfExercise models.PDFExercise
		if err := db.Preload("Creator").First(&pdfExercise, "material_id = ?", referenceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return material.ToJSON(), nil
			}
			return nil, fmt.Errorf("failed to get PDF exercise: %w", err)
		}
		return pdfExercise.ToJSON(), nil

	case "video":
		var video models.Video
		if err := db.Preload("Creator").First(&video, "material_id = ?", referenceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return material.ToJSON(), nil
			}
			return nil, fmt.Errorf("failed to get video: %w", err)
		}
		return video.ToJSON(), nil

	case "document":
		var document models.Document
		if err := db.Preload("Creator").First(&document, "material_id = ?", referenceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return material.ToJSON(), nil
			}
			return nil, fmt.Errorf("failed to get document: %w", err)
		}
		return document.ToJSON(), nil

	case "announcement":
		var announcement models.Announcement
		if err := db.Preload("Creator").First(&announcement, "material_id = ?", referenceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return material.ToJSON(), nil
			}
			return nil, fmt.Errorf("failed to get announcement: %w", err)
		}
		return announcement.ToJSON(), nil

	default:
		// Unknown reference type, return basic CourseMaterial data
		return material.ToJSON(), nil
	}
}

















