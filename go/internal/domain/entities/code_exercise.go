package models

import (
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/week"
	"gorm.io/gorm"
)

// CodeExercise represents a code exercise material
type CodeExercise struct {
	MaterialBase
	TotalPoints      *int           `json:"total_points,omitempty" gorm:"type:int;not null"`
	Deadline         *string        `json:"deadline,omitempty" gorm:"type:varchar(50)"`
	IsGraded         *bool          `json:"is_graded,omitempty" gorm:"type:boolean;default:true"`
	ProblemStatement string         `json:"problem_statement,omitempty" gorm:"type:text;not null"`
	ProblemImages    types.JSONData `json:"problem_images,omitempty" gorm:"type:jsonb;default:'[]'::jsonb"`
	ExampleInputs    types.JSONData `json:"example_inputs,omitempty" gorm:"type:jsonb;default:'[]'::jsonb"`
	ExampleOutputs   types.JSONData `json:"example_outputs,omitempty" gorm:"type:jsonb;default:'[]'::jsonb"`
	Constraints      string         `json:"constraints,omitempty" gorm:"type:text"`
	Hints            string         `json:"hints,omitempty" gorm:"type:text"`

	// Relations
	TestCases []TestCase `json:"test_cases,omitempty" gorm:"foreignKey:MaterialID;references:MaterialID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName returns the table name
func (CodeExercise) TableName() string {
	return "code_exercises"
}

// GetMaterialType returns the material type
func (ce *CodeExercise) GetMaterialType() string {
	return string(enums.MaterialTypeCodeExercise)
}

// ToJSON converts CodeExercise to JSON map
func (ce *CodeExercise) ToJSON() map[string]interface{} {
	result := ce.MaterialBase.ToJSONBase()
	result["type"] = enums.MaterialTypeCodeExercise
	result["submission_type"] = "code"

	if ce.TotalPoints != nil {
		result["total_points"] = *ce.TotalPoints
	}
	if ce.Deadline != nil {
		result["deadline"] = *ce.Deadline
	}
	if ce.IsGraded != nil {
		result["is_graded"] = *ce.IsGraded
	}
	if ce.ProblemStatement != "" {
		result["problem_statement"] = ce.ProblemStatement
	}
	if ce.ProblemImages != nil {
		result["problem_images"] = ce.ProblemImages
	}
	if ce.ExampleInputs != nil {
		result["example_inputs"] = ce.ExampleInputs
	}
	if ce.ExampleOutputs != nil {
		result["example_outputs"] = ce.ExampleOutputs
	}
	if ce.Constraints != "" {
		result["constraints"] = ce.Constraints
	}
	if ce.Hints != "" {
		result["hints"] = ce.Hints
	}

	if ce.Creator.UserID != "" {
		result["creator"] = ce.Creator.ToJSON()
	}

	// Add test cases if present
	if len(ce.TestCases) > 0 {
		testCasesJSON := make([]map[string]interface{}, len(ce.TestCases))
		for i, testCase := range ce.TestCases {
			testCasesJSON[i] = testCase.ToJSON()
		}
		result["test_cases"] = testCasesJSON
	}

	return result
}

// BeforeCreate sets the material ID if not already set
func (ce *CodeExercise) BeforeCreate(tx *gorm.DB) error {
	return ce.MaterialBase.BeforeCreate(tx)
}

// IsCodeExercise returns true
func (ce *CodeExercise) IsCodeExercise() bool {
	return true
}

// IsExercise returns true
func (ce *CodeExercise) IsExercise() bool {
	return true
}

// RequiresSubmission returns true
func (ce *CodeExercise) RequiresSubmission() bool {
	return true
}

// RequiresApproval returns false (code exercises are auto-graded)
func (ce *CodeExercise) RequiresApproval() bool {
	return false
}

// HasTestCases returns true if test cases exist
func (ce *CodeExercise) HasTestCases() bool {
	return len(ce.TestCases) > 0
}

// GetTestCasesCount returns the number of test cases
func (ce *CodeExercise) GetTestCasesCount() int {
	return len(ce.TestCases)
}

// CanAddTestCases returns true
func (ce *CodeExercise) CanAddTestCases() bool {
	return true
}

// CanSubmitCode returns true if test cases exist
func (ce *CodeExercise) CanSubmitCode() bool {
	return ce.HasTestCases()
}

// RequiresStrictDeadline returns true if graded, false if practice
func (ce *CodeExercise) RequiresStrictDeadline() bool {
	if ce.IsGraded != nil && !*ce.IsGraded {
		return false
	}
	return true
}

// WeekBasedEntity interface implementation
func (ce *CodeExercise) GetWeek() int {
	return ce.MaterialBase.Week
}

func (ce *CodeExercise) SetWeek(week int) {
	ce.MaterialBase.Week = week
}

func (ce *CodeExercise) GetTableName() string {
	return "code_exercises"
}

func (ce *CodeExercise) GetCourseID() string {
	return ce.MaterialBase.CourseID
}

// Ensure CodeExercise implements WeekBasedEntity interface
var _ week.WeekBasedEntity = (*CodeExercise)(nil)
