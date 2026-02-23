package material

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Project-DSView/backend/go/internal/domain/enums"
)

// ValidateMaterialType validates if a material type is valid
func ValidateMaterialType(materialType string) bool {
	switch enums.MaterialType(materialType) {
	case enums.MaterialTypeDocument, enums.MaterialTypeVideo, enums.MaterialTypeCodeExercise, enums.MaterialTypePDFExercise:
		return true
	default:
		return false
	}
}

// ValidateVideoFields validates required fields for video material
func ValidateVideoFields(videoURL string) error {
	if strings.TrimSpace(videoURL) == "" {
		return errors.New("video_url is required for video material")
	}
	return nil
}

// ValidateDocumentFields validates required fields for document material
func ValidateDocumentFields(fileURL, fileName string) error {
	if strings.TrimSpace(fileURL) == "" {
		return errors.New("file_url is required for document material")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("file_name is required for document material")
	}
	return nil
}

// ValidateCodeExerciseFields validates required fields for code exercise
func ValidateCodeExerciseFields(totalPoints *int, problemStatement string) error {
	if totalPoints == nil || *totalPoints <= 0 {
		return errors.New("total_points must be greater than 0 for code exercise")
	}
	if strings.TrimSpace(problemStatement) == "" {
		return errors.New("problem_statement is required for code exercise")
	}
	return nil
}

// ValidatePDFExerciseFields validates required fields for PDF exercise
func ValidatePDFExerciseFields(totalPoints *int, fileURL, fileName string) error {
	if totalPoints == nil || *totalPoints <= 0 {
		return errors.New("total_points must be greater than 0 for PDF exercise")
	}
	if strings.TrimSpace(fileURL) == "" {
		return errors.New("file_url is required for PDF exercise")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("file_name is required for PDF exercise")
	}
	return nil
}

// ValidateMaterialFields validates fields based on material type
func ValidateMaterialFields(materialType string, fields map[string]interface{}) error {
	switch enums.MaterialType(materialType) {
	case enums.MaterialTypeVideo:
		videoURL, _ := fields["video_url"].(string)
		return ValidateVideoFields(videoURL)

	case enums.MaterialTypeDocument:
		fileURL, _ := fields["file_url"].(string)
		fileName, _ := fields["file_name"].(string)
		return ValidateDocumentFields(fileURL, fileName)

	case enums.MaterialTypeCodeExercise:
		totalPoints, _ := fields["total_points"].(*int)
		problemStatement, _ := fields["problem_statement"].(string)
		return ValidateCodeExerciseFields(totalPoints, problemStatement)

	case enums.MaterialTypePDFExercise:
		totalPoints, _ := fields["total_points"].(*int)
		fileURL, _ := fields["file_url"].(string)
		fileName, _ := fields["file_name"].(string)
		return ValidatePDFExerciseFields(totalPoints, fileURL, fileName)

	default:
		return fmt.Errorf("unknown material type: %s", materialType)
	}
}


















