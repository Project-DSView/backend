package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MaterialBase contains common fields shared by all material types
type MaterialBase struct {
	MaterialID  string    `json:"material_id" gorm:"primaryKey;type:varchar(36)"`
	CourseID    string    `json:"course_id" gorm:"type:varchar(36);not null;index"`
	Title       string    `json:"title" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	Week        int       `json:"week" gorm:"type:int;not null;default:0;index"`
	IsPublic    bool      `json:"is_public" gorm:"default:true;not null"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(36);not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Course  Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:CourseID"`
	Creator User   `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:UserID"`
}

// Material interface defines common operations for all material types
type Material interface {
	GetMaterialID() string
	GetCourseID() string
	GetTitle() string
	GetDescription() string
	GetWeek() int
	GetIsPublic() bool
	GetCreatedBy() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetMaterialType() string
	ToJSON() map[string]interface{}
	BeforeCreate(tx *gorm.DB) error
	TableName() string
}

// GetMaterialID returns the material ID
func (mb *MaterialBase) GetMaterialID() string {
	return mb.MaterialID
}

// GetCourseID returns the course ID
func (mb *MaterialBase) GetCourseID() string {
	return mb.CourseID
}

// GetTitle returns the title
func (mb *MaterialBase) GetTitle() string {
	return mb.Title
}

// GetDescription returns the description
func (mb *MaterialBase) GetDescription() string {
	return mb.Description
}

// GetWeek returns the week number
func (mb *MaterialBase) GetWeek() int {
	return mb.Week
}

// GetIsPublic returns whether the material is public
func (mb *MaterialBase) GetIsPublic() bool {
	return mb.IsPublic
}

// GetCreatedBy returns the creator user ID
func (mb *MaterialBase) GetCreatedBy() string {
	return mb.CreatedBy
}

// GetCreatedAt returns the creation time
func (mb *MaterialBase) GetCreatedAt() time.Time {
	return mb.CreatedAt
}

// GetUpdatedAt returns the update time
func (mb *MaterialBase) GetUpdatedAt() time.Time {
	return mb.UpdatedAt
}

// BeforeCreate sets the material ID if not already set
func (mb *MaterialBase) BeforeCreate(tx *gorm.DB) error {
	if mb.MaterialID == "" {
		mb.MaterialID = uuid.New().String()
	}
	return nil
}

// ToJSONBase returns the base JSON representation
func (mb *MaterialBase) ToJSONBase() map[string]interface{} {
	return map[string]interface{}{
		"material_id": mb.MaterialID,
		"course_id":   mb.CourseID,
		"title":       mb.Title,
		"description": mb.Description,
		"week":        mb.Week,
		"is_public":    mb.IsPublic,
		"created_by":   mb.CreatedBy,
		"created_at":   mb.CreatedAt,
		"updated_at":   mb.UpdatedAt,
	}
}


















