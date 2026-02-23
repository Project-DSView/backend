package material

import (
	"errors"
	"gorm.io/gorm"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
)

// CheckMaterialCreatorPermission checks if a user is the creator of a material
func CheckMaterialCreatorPermission(db *gorm.DB, materialID, userID string, materialType string) (bool, error) {
	var createdBy string
	var err error

	switch materialType {
	case "video":
		var video models.Video
		if err = db.First(&video, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		createdBy = video.CreatedBy
	case "document":
		var doc models.Document
		if err = db.First(&doc, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		createdBy = doc.CreatedBy
	case "code_exercise":
		var ex models.CodeExercise
		if err = db.First(&ex, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		createdBy = ex.CreatedBy
	case "pdf_exercise":
		var ex models.PDFExercise
		if err = db.First(&ex, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		createdBy = ex.CreatedBy
	default:
		return false, errors.New("unknown material type")
	}

	return createdBy == userID, nil
}

// CheckTeacherPermission checks if a user is a teacher
func CheckTeacherPermission(db *gorm.DB, userID string) (bool, error) {
	var user models.User
	if err := db.First(&user, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("user not found")
		}
		return false, err
	}
	return user.IsTeacher, nil
}

// ValidateMaterialAccess validates if a user can access a material
func ValidateMaterialAccess(db *gorm.DB, materialID, userID string, materialType string, isTeacher bool) (bool, error) {
	// Teachers can always access
	if isTeacher {
		return true, nil
	}

	// Check if material is public
	var isPublic bool
	var err error

	switch materialType {
	case "video":
		var video models.Video
		if err = db.First(&video, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		isPublic = video.IsPublic
	case "document":
		var doc models.Document
		if err = db.First(&doc, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		isPublic = doc.IsPublic
	case "code_exercise":
		var ex models.CodeExercise
		if err = db.First(&ex, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		isPublic = ex.IsPublic
	case "pdf_exercise":
		var ex models.PDFExercise
		if err = db.First(&ex, "material_id = ?", materialID).Error; err != nil {
			return false, err
		}
		isPublic = ex.IsPublic
	default:
		return false, errors.New("unknown material type")
	}

	return isPublic, nil
}


















