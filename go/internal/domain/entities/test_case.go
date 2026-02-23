package models

import (
	"time"

	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TestCase struct {
	TestCaseID     string         `json:"test_case_id" gorm:"primaryKey;type:varchar(36)"`
	MaterialID     *string        `json:"material_id,omitempty" gorm:"type:varchar(36);index"` // For course material test cases
	MaterialType   string         `json:"material_type,omitempty" gorm:"type:varchar(20);index"` // Polymorphic: code_exercise (only code exercises have test cases)
	InputData      types.JSONData `json:"input_data" gorm:"type:jsonb;not null"`
	ExpectedOutput types.JSONData `json:"expected_output" gorm:"type:jsonb;not null"`
	IsPublic       bool           `json:"is_public" gorm:"default:false;not null"`         // Whether this test case is visible to students
	DisplayName    string         `json:"display_name,omitempty" gorm:"type:varchar(255)"` // Human-readable name for the test case
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	CourseMaterial *CourseMaterial `json:"course_material,omitempty" gorm:"foreignKey:MaterialID;references:MaterialID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (tc *TestCase) BeforeCreate(tx *gorm.DB) error {
	if tc.TestCaseID == "" {
		tc.TestCaseID = uuid.New().String()
	}
	return nil
}

// TableName overrides the table name used by TestCase to `test_cases`
func (TestCase) TableName() string {
	return "test_cases"
}

// ToJSON returns a map representation suitable for JSON responses
func (tc *TestCase) ToJSON() map[string]interface{} {
	// Convert JSONData to string for input_data and expected_output
	var inputDataStr string
	var expectedOutputStr string
	if tc.InputData != nil {
		inputDataStr = string(tc.InputData)
	}
	if tc.ExpectedOutput != nil {
		expectedOutputStr = string(tc.ExpectedOutput)
	}

	result := map[string]interface{}{
		"test_case_id":    tc.TestCaseID,
		"input_data":      inputDataStr,
		"expected_output": expectedOutputStr,
		"is_public":       tc.IsPublic,
		"display_name":    tc.DisplayName,
		"created_at":      tc.CreatedAt,
		"updated_at":      tc.UpdatedAt,
	}

	// Add material_id if present (for new system)
	if tc.MaterialID != nil {
		result["material_id"] = *tc.MaterialID
	}

	// Add course material relation if present
	if tc.CourseMaterial != nil && tc.CourseMaterial.MaterialID != "" {
		result["course_material"] = tc.CourseMaterial.ToJSON()
	}

	return result
}

// Helper methods for test case types
func (tc *TestCase) IsCodeExercise() bool {
	return tc.MaterialID != nil && tc.CourseMaterial != nil && tc.CourseMaterial.IsCodeExercise()
}

func (tc *TestCase) IsPDFExercise() bool {
	return tc.MaterialID != nil && tc.CourseMaterial != nil && tc.CourseMaterial.IsPDFExercise()
}

func (tc *TestCase) RequiresAutoGrading() bool {
	return tc.IsCodeExercise()
}

func (tc *TestCase) RequiresManualGrading() bool {
	return tc.IsPDFExercise()
}
