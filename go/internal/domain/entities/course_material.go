package models

import (
	"time"

	"github.com/Project-DSView/backend/go/internal/domain/enums"
	"github.com/Project-DSView/backend/go/pkg/week"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CourseMaterial is a central table that references specific material tables
// It uses polymorphic association with reference_id and reference_type
type CourseMaterial struct {
	MaterialID    string             `json:"material_id" gorm:"primaryKey;type:varchar(36)"`
	CourseID      string             `json:"course_id" gorm:"type:varchar(36);not null;index"`
	Type          enums.MaterialType `json:"type" gorm:"type:varchar(20);not null;index"`
	Week          int                `json:"week" gorm:"type:int;not null;default:0;index"`
	ReferenceID   *string            `json:"reference_id,omitempty" gorm:"type:varchar(36);index"`   // Polymorphic reference ID
	ReferenceType *string            `json:"reference_type,omitempty" gorm:"type:varchar(20);index"` // Polymorphic reference type: 'video', 'document', 'code_exercise', 'pdf_exercise', 'announcement'
	CreatedAt     time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time          `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Course Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`
}

func (cm *CourseMaterial) BeforeCreate(tx *gorm.DB) error {
	if cm.MaterialID == "" {
		cm.MaterialID = uuid.New().String()
	}
	return nil
}

func (CourseMaterial) TableName() string {
	return "course_materials"
}

func (cm *CourseMaterial) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"material_id": cm.MaterialID,
		"course_id":   cm.CourseID,
		"type":        cm.Type,
		"week":        cm.Week,
		"created_at":  cm.CreatedAt,
		"updated_at":  cm.UpdatedAt,
	}

	if cm.ReferenceID != nil {
		result["reference_id"] = *cm.ReferenceID
	}
	if cm.ReferenceType != nil {
		result["reference_type"] = *cm.ReferenceType
	}

	return result
}

func IsValidMaterialType(materialType string) bool {
	switch enums.MaterialType(materialType) {
	case enums.MaterialTypeAnnouncement, enums.MaterialTypeDocument, enums.MaterialTypeVideo, enums.MaterialTypeCodeExercise, enums.MaterialTypePDFExercise:
		return true
	default:
		return false
	}
}

// Helper methods for material types
func (cm *CourseMaterial) IsCodeExercise() bool {
	return cm.Type == enums.MaterialTypeCodeExercise
}

func (cm *CourseMaterial) IsPDFExercise() bool {
	return cm.Type == enums.MaterialTypePDFExercise
}

func (cm *CourseMaterial) IsExercise() bool {
	return cm.IsCodeExercise() || cm.IsPDFExercise()
}

func (cm *CourseMaterial) IsAnnouncement() bool {
	return cm.Type == enums.MaterialTypeAnnouncement
}

func (cm *CourseMaterial) IsVideo() bool {
	return cm.Type == enums.MaterialTypeVideo
}

func (cm *CourseMaterial) IsDocument() bool {
	return cm.Type == enums.MaterialTypeDocument
}

// GetReferenceID returns the reference ID if set
func (cm *CourseMaterial) GetReferenceID() string {
	if cm.ReferenceID != nil {
		return *cm.ReferenceID
	}
	return ""
}

// GetReferenceType returns the reference type if set
func (cm *CourseMaterial) GetReferenceType() string {
	if cm.ReferenceType != nil {
		return *cm.ReferenceType
	}
	return ""
}

// WeekBasedEntity interface implementation
func (cm *CourseMaterial) GetWeek() int {
	return cm.Week
}

func (cm *CourseMaterial) SetWeek(week int) {
	cm.Week = week
}

func (cm *CourseMaterial) GetTableName() string {
	return "course_materials"
}

func (cm *CourseMaterial) GetCourseID() string {
	return cm.CourseID
}

// Ensure CourseMaterial implements WeekBasedEntity interface
var _ week.WeekBasedEntity = (*CourseMaterial)(nil)
