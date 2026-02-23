package material

import (
	"errors"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"gorm.io/gorm"
)

// GetMaterialCreator gets the creator user ID from the actual material table
func GetMaterialCreator(db *gorm.DB, referenceID, referenceType string) (string, error) {
	switch referenceType {
	case "video":
		var video models.Video
		if err := db.First(&video, "material_id = ?", referenceID).Error; err != nil {
			return "", err
		}
		return video.CreatedBy, nil
	case "document":
		var document models.Document
		if err := db.First(&document, "material_id = ?", referenceID).Error; err != nil {
			return "", err
		}
		return document.CreatedBy, nil
	case "code_exercise":
		var codeExercise models.CodeExercise
		if err := db.First(&codeExercise, "material_id = ?", referenceID).Error; err != nil {
			return "", err
		}
		return codeExercise.CreatedBy, nil
	case "pdf_exercise":
		var pdfExercise models.PDFExercise
		if err := db.First(&pdfExercise, "material_id = ?", referenceID).Error; err != nil {
			return "", err
		}
		return pdfExercise.CreatedBy, nil
	case "announcement":
		var announcement models.Announcement
		if err := db.First(&announcement, "material_id = ?", referenceID).Error; err != nil {
			return "", err
		}
		return announcement.CreatedBy, nil
	default:
		return "", errors.New("invalid reference type")
	}
}

// GetMaterialFileURL gets the file URL from the actual material table (for documents and PDF exercises)
func GetMaterialFileURL(db *gorm.DB, referenceID, referenceType string) (string, error) {
	switch referenceType {
	case "document":
		var document models.Document
		if err := db.First(&document, "material_id = ?", referenceID).Error; err != nil {
			return "", err
		}
		return document.FileURL, nil
	case "pdf_exercise":
		var pdfExercise models.PDFExercise
		if err := db.First(&pdfExercise, "material_id = ?", referenceID).Error; err != nil {
			return "", err
		}
		return pdfExercise.FileURL, nil
	default:
		return "", nil // Other types don't have file URLs
	}
}

// VerifyReferenceExists verifies that the referenced material exists in the appropriate table
func VerifyReferenceExists(db *gorm.DB, referenceID, referenceType string) error {
	switch referenceType {
	case "video":
		var video models.Video
		if err := db.First(&video, "material_id = ?", referenceID).Error; err != nil {
			return errors.New("referenced video not found")
		}
	case "document":
		var document models.Document
		if err := db.First(&document, "material_id = ?", referenceID).Error; err != nil {
			return errors.New("referenced document not found")
		}
	case "code_exercise":
		var codeExercise models.CodeExercise
		if err := db.First(&codeExercise, "material_id = ?", referenceID).Error; err != nil {
			return errors.New("referenced code exercise not found")
		}
	case "pdf_exercise":
		var pdfExercise models.PDFExercise
		if err := db.First(&pdfExercise, "material_id = ?", referenceID).Error; err != nil {
			return errors.New("referenced PDF exercise not found")
		}
	case "announcement":
		var announcement models.Announcement
		if err := db.First(&announcement, "material_id = ?", referenceID).Error; err != nil {
			return errors.New("referenced announcement not found")
		}
	default:
		return errors.New("invalid reference type")
	}
	return nil
}


















