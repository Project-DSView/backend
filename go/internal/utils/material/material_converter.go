package material

import (
	"fmt"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
)

// ConvertToMaterial converts a material type to the Material interface
func ConvertToMaterial(material interface{}) (models.Material, error) {
	switch m := material.(type) {
	case *models.Video:
		return m, nil
	case *models.CodeExercise:
		return m, nil
	case *models.PDFExercise:
		return m, nil
	case *models.Document:
		return m, nil
	default:
		return nil, fmt.Errorf("unknown material type: %T", material)
	}
}

// GetMaterialType returns the material type string from a material instance
func GetMaterialType(material models.Material) string {
	return material.GetMaterialType()
}

// ConvertFromOldCourseMaterial converts old CourseMaterial to new material types
// NOTE: This function is deprecated. The current CourseMaterial model uses polymorphic references
// and does not contain the material-specific fields. This function returns materials with zero values.
func ConvertFromOldCourseMaterial(oldMaterial *models.CourseMaterial) (models.Material, error) {
	base := models.MaterialBase{
		MaterialID:  oldMaterial.MaterialID,
		CourseID:    oldMaterial.CourseID,
		Title:       "", // CourseMaterial no longer has Title field
		Description: "", // CourseMaterial no longer has Description field
		Week:        oldMaterial.Week,
		IsPublic:    true, // CourseMaterial no longer has IsPublic field, defaulting to true
		CreatedBy:   "",   // CourseMaterial no longer has CreatedBy field
		CreatedAt:   oldMaterial.CreatedAt,
		UpdatedAt:   oldMaterial.UpdatedAt,
	}

	switch oldMaterial.Type {
	case enums.MaterialTypeVideo:
		return &models.Video{
			MaterialBase: base,
			VideoURL:     "", // CourseMaterial no longer has VideoURL field
		}, nil

	case enums.MaterialTypeDocument:
		return &models.Document{
			MaterialBase: base,
			FileURL:      "", // CourseMaterial no longer has FileURL field
			FileName:     "", // CourseMaterial no longer has FileName field
			FileSize:     0,  // CourseMaterial no longer has FileSize field
			MimeType:     "", // CourseMaterial no longer has MimeType field
		}, nil

	case enums.MaterialTypeCodeExercise:
		return &models.CodeExercise{
			MaterialBase:     base,
			TotalPoints:      nil, // CourseMaterial no longer has TotalPoints field
			Deadline:         nil, // CourseMaterial no longer has Deadline field
			IsGraded:         nil, // CourseMaterial no longer has IsGraded field
			ProblemStatement: "",  // CourseMaterial no longer has ProblemStatement field
			ProblemImages:    nil, // CourseMaterial no longer has ProblemImages field
			ExampleInputs:    nil, // CourseMaterial no longer has ExampleInputs field
			ExampleOutputs:   nil, // CourseMaterial no longer has ExampleOutputs field
			Constraints:      "",  // CourseMaterial no longer has Constraints field
			Hints:            "",  // CourseMaterial no longer has Hints field
		}, nil

	case enums.MaterialTypePDFExercise:
		return &models.PDFExercise{
			MaterialBase: base,
			TotalPoints:  nil, // CourseMaterial no longer has TotalPoints field
			Deadline:     nil, // CourseMaterial no longer has Deadline field
			IsGraded:     nil, // CourseMaterial no longer has IsGraded field
			FileURL:      "",  // CourseMaterial no longer has FileURL field
			FileName:     "",  // CourseMaterial no longer has FileName field
			FileSize:     0,   // CourseMaterial no longer has FileSize field
			MimeType:     "",  // CourseMaterial no longer has MimeType field
		}, nil

	default:
		return nil, fmt.Errorf("unknown material type: %s", oldMaterial.Type)
	}
}
