package models

import (
	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/week"
	"gorm.io/gorm"
)

// PDFExercise represents a PDF exercise material
type PDFExercise struct {
	MaterialBase
	TotalPoints *int    `json:"total_points,omitempty" gorm:"type:int;not null"`
	Deadline    *string `json:"deadline,omitempty" gorm:"type:varchar(50)"`
	IsGraded     *bool   `json:"is_graded,omitempty" gorm:"type:boolean;default:true"`
	FileURL     string  `json:"file_url" gorm:"type:text;not null"`
	FileName    string  `json:"file_name" gorm:"type:varchar(255);not null"`
	FileSize    int64   `json:"file_size" gorm:"type:bigint;default:0"`
	MimeType    string  `json:"mime_type" gorm:"type:varchar(100)"`
}

// TableName returns the table name
func (PDFExercise) TableName() string {
	return "pdf_exercises"
}

// GetMaterialType returns the material type
func (pe *PDFExercise) GetMaterialType() string {
	return string(enums.MaterialTypePDFExercise)
}

// ToJSON converts PDFExercise to JSON map
func (pe *PDFExercise) ToJSON() map[string]interface{} {
	result := pe.MaterialBase.ToJSONBase()
	result["type"] = enums.MaterialTypePDFExercise
	result["submission_type"] = "file"

	if pe.TotalPoints != nil {
		result["total_points"] = *pe.TotalPoints
	}
	if pe.Deadline != nil {
		result["deadline"] = *pe.Deadline
	}
	if pe.IsGraded != nil {
		result["is_graded"] = *pe.IsGraded
	}
	result["file_url"] = pe.FileURL
	result["file_name"] = pe.FileName
	result["file_size"] = pe.FileSize
	result["mime_type"] = pe.MimeType

	if pe.Creator.UserID != "" {
		result["creator"] = pe.Creator.ToJSON()
	}

	return result
}

// BeforeCreate sets the material ID if not already set
func (pe *PDFExercise) BeforeCreate(tx *gorm.DB) error {
	return pe.MaterialBase.BeforeCreate(tx)
}

// IsPDFExercise returns true
func (pe *PDFExercise) IsPDFExercise() bool {
	return true
}

// IsExercise returns true
func (pe *PDFExercise) IsExercise() bool {
	return true
}

// RequiresSubmission returns true
func (pe *PDFExercise) RequiresSubmission() bool {
	return true
}

// RequiresApproval returns true (PDF exercises need manual approval)
func (pe *PDFExercise) RequiresApproval() bool {
	return true
}

// HasFile returns true if file URL is set
func (pe *PDFExercise) HasFile() bool {
	return pe.FileURL != "" && pe.FileName != ""
}

// GetDisplayURL returns the file URL
func (pe *PDFExercise) GetDisplayURL() string {
	return pe.FileURL
}

// CanSubmitFile returns true
func (pe *PDFExercise) CanSubmitFile() bool {
	return true
}

// RequiresStrictDeadline returns true if graded, false if practice
func (pe *PDFExercise) RequiresStrictDeadline() bool {
	if pe.IsGraded != nil && !*pe.IsGraded {
		return false
	}
	return true
}

// WeekBasedEntity interface implementation
func (pe *PDFExercise) GetWeek() int {
	return pe.MaterialBase.Week
}

func (pe *PDFExercise) SetWeek(week int) {
	pe.MaterialBase.Week = week
}

func (pe *PDFExercise) GetTableName() string {
	return "pdf_exercises"
}

func (pe *PDFExercise) GetCourseID() string {
	return pe.MaterialBase.CourseID
}

// Ensure PDFExercise implements WeekBasedEntity interface
var _ week.WeekBasedEntity = (*PDFExercise)(nil)


















